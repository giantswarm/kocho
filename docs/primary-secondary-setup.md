# Setting up a Primary-Secondary cluster with Kocho

Kocho comes with out-of-the-box support for setting up a primary-secondary cluster, with Etcd and fleet.

This configuration sets up a primary Etcd cluster, and a secondary cluster that sets Etcd to proxy from the primary cluster. Setting up a cluster like this allows for more nodes to be utilised overall, without overloading Etcd.
See [the Etcd proxy documentation](https://coreos.com/etcd/docs/latest/proxy.html) for more information.

## Set up

First of all, we'll detail the set up, just to avoid any confusion.

The `kocho.yml` configuration file we're using looks as follows. Some data has had to be censored for security reasons - any key with the value `--` needs to be set with a value from your setup.

```yaml
# Configure kocho
cluster-size: 3
certificate: --

machine-type: t2.micro

aws-vpc: --
aws-vpc-cidr: --
aws-keypair: --
aws-subnet: --
aws-az: eu-west-1a

# DNS
dns-zone: --
```

We have `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` environment variables set in the environment, from IAM credentials.

## Setting up templates

The command

```nohighlight
$ kocho template-init
```

will copy the default templates out of the binary itself into the `templates` directory. The default templates are set up for starting a basic primary-secondary cluster.

## Creating the primary cluster

To create the primary cluster, use a command like this:

```nohighlight
$ kocho create --type=primary batman
```

If you look at your AWS CloudFormation control panel, you should see that the stack has been created with the name `batman`.

## Inspecting the primary cluster

If we log into one of the AWS EC2 instances that have been brought up by the auto scaling group via SSH, we should see something like this:

```nohighlight
CoreOS stable (681.2.0)
Update Strategy: No Reboots
```

The `etcdctl cluster-health` command should show us roughly this

```nohighlight
cluster is healthy
member 2845210c76c93874 is healthy
member 5e5ffea7c416c645 is healthy
member f0e9d438535c2b86 is healthy
```

and `fleetctl list-machines` should show our machines similar to the following output:

```nohighlight
MACHINE     IP              METADATA
00fa83e0... 172.31.19.121   role=primary,role-core=true
56c031ea... 172.31.18.18    role=primary,role-core=true
7b977366... 172.31.26.67    role=primary,role-core=true
```

As you can see, we have three nodes in our Etcd cluster, and some useful fleet metadata has been set.

## Creating the secondary cluster

To create the secondary cluster, we'll need to get both the Etcd discovery URL that the primary cluster used, as well as a list of Etcd peers. Kocho can get both for you:

```nohighlight
$ kocho etcd discovery batman
https://discovery.etcd.io/d5a0e0819e201c8103d346df8b20ed55
```

```nohighlight
$ kocho etcd peers batman
http://172.31.18.18:2379,http://172.31.26.67:2379,http://172.31.19.121:2379
```

We can then use this information to create the secondary cluster:

```nohighlight
$ kocho create \
  --type=secondary \
  --etcd-discovery-url=https://discovery.etcd.io/d5a0e0819e201c8103d346df8b20ed55 \
  --etcd-peers=http://172.31.18.18:2379,http://172.31.26.67:2379,http://172.31.19.121:2379 \
  robin
```

Like before, inspecting AWS CloudFormation and AWS EC2 control panels show that the cluster has been set up correctly. There should now be two AWS CloudFormation stacks, each with three AWS EC2 instances.

## Inspecting the secondary cluster

We log in via SSH to one of the instances of the secondary cluster and check if everything is set up properly:

```nohighlight
CoreOS stable (681.2.0)
Update Strategy: No Reboots

core@ip-172-31-17-39 ~ $ etcdctl cluster-health
cluster is healthy
member 2845210c76c93874 is healthy
member 5e5ffea7c416c645 is healthy
member f0e9d438535c2b86 is healthy

core@ip-172-31-17-39 ~ $ etcdctl member list
2845210c76c93874: name=7b977366e39a48e6ab9edd132472bae0 peerURLs=http://172.31.26.67:2380 clientURLs=http://172.31.26.67:2379
5e5ffea7c416c645: name=00fa83e0c7d44e99845ed53f11e56531 peerURLs=http://172.31.19.121:2380 clientURLs=http://172.31.19.121:2379
f0e9d438535c2b86: name=56c031ea78ee4d0e852cd0ea0333d657 peerURLs=http://172.31.18.18:2380 clientURLs=http://172.31.18.18:2379

core@ip-172-31-17-39 ~ $ fleetctl list-machines
MACHINE		IP		METADATA
00fa83e0...	172.31.19.121	role=primary,role-core=true
2b3cf72b...	172.31.17.39	role=secondary,role-worker=true,stack-compute=true
56c031ea...	172.31.18.18	role=primary,role-core=true
7b977366...	172.31.26.67	role=primary,role-core=true
888dbc95...	172.31.17.40	role=secondary,role-worker=true,stack-compute=true
ec5c4304...	172.31.17.41	role=secondary,role-worker=true,stack-compute=true
```

We can see that the secondary cluster is set up correctly, and is using the primary Etcd nodes. Furthermore, we have set the fleet metadata differently for the secondary cluster.

## Listing all clusters:

We can view both clusters, as well as their type, from Kocho itself:

```nohighlight
$ kocho list
Name         Type        Created
robin        secondary   11 Feb 16 16:48 UTC
batman       primary     11 Feb 16 16:40 UTC
```

## Cleanup

If you've been following along, the following will remove both clusters completely:

```nohighlight
$ kocho destroy batman && kocho destroy robin
```
