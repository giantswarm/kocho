# Setting up a Primary-Secondary cluster with Kocho

`kocho` comes with out-of-the-box support for setting up a primary-secondary cluster, with `etcd` and `fleet`.

This configuration sets up a primary `etcd` cluster, and a secondary cluster that sets `etcd` to proxy from the primary cluster.
Setting up a cluster like this allows for more nodes to be utilised overall without overloading `etcd`.

## Set up

First of all, I'll detail my set up, just to avoid any confusion:
  - The config file I'm using looks as follows. I've had to censor some data for security reasons - any key with the value "--" needs to be set with a value from your set up.
    ```
    $ cat ./kocho.yml
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
  - I have `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` environment variables set in my environment, from my IAM credentials.
  
## Setting up templates
```
$ kocho template-init
```
This will copy the default templates out of the binary itself, into the directory `templates/`. The default templates are set up for starting a basic primary-secondary cluster.

## Creating the primary cluster
```
$ kocho create --type=primary batman
```
If you look at your AWS CloudFormation control panel, you should see the stack has been created, with the name `batman`.

## Inspecting the primary cluster

I'm going to SSH into one of the AWS EC2 instances that have been brought up by the auto scaling group:
```
CoreOS stable (681.2.0)
Update Strategy: No Reboots

core@ip-172-31-18-18 ~ $ etcdctl cluster-health
cluster is healthy
member 2845210c76c93874 is healthy
member 5e5ffea7c416c645 is healthy
member f0e9d438535c2b86 is healthy

core@ip-172-31-18-18 ~ $ fleetctl list-machines
MACHINE		IP		METADATA
00fa83e0...	172.31.19.121	role=primary,role-core=true
56c031ea...	172.31.18.18	role=primary,role-core=true
7b977366...	172.31.26.67	role=primary,role-core=true
```

As you can see, we have three nodes in an `etcd` cluster, and some useful `fleet` metadata has been set.

## Creating the secondary cluster

To create the secondary cluster, we'll need to get both the `etcd` discovery URL that the primary cluster used, as well as a list of `etcd` peers - Kocho can do this for you.

```
$ kocho etcd discovery batman
https://discovery.etcd.io/d5a0e0819e201c8103d346df8b20ed55
$ kocho etcd peers batman
http://172.31.18.18:2379,http://172.31.26.67:2379,http://172.31.19.121:2379
```

We can then use this information to create the secondary cluster:
```
$ kocho create --type=secondary \
--etcd-discovery-url=https://discovery.etcd.io/d5a0e0819e201c8103d346df8b20ed55 \
--etcd-peers=http://172.31.18.18:2379,http://172.31.26.67:2379,http://172.31.19.121:2379 \
robin
```

Like before, inspecting AWS CloudFormation and AWS EC2 control panels show that the cluster has been set up correctly. There should now be 2 AWS CloudFormation stacks, each with 3 AWS EC2 instances.

## Inspecting the secondary cluster

I'm going to SSH into one of the instances of the secondary cluster:
```
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

We can see that the secondary cluster is set up correctly, and is using the primary `etcd` nodes. Furthermore, we have set the `fleet` metadata differently for the secondary cluster.

## Listing all clusters:

We can view both clusters, as well as their type, from Kocho itself.
```
$ kocho list
Name         Type        Created
robin        secondary   11 Feb 16 16:48 UTC
batman       primary     11 Feb 16 16:40 UTC
```

## Cleanup

If you've been following along, the following will remove both clusters completely:
```
$ kocho destroy batman && kocho destroy robin

```

Thanks! <3