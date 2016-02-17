# Getting Started with Kocho

Kocho is a tool around AWS CloudFormation and CloudFlare to automate the setup
of CoreOS clusters with customized versions of Etcd, Fleet and Docker.

## Prerequisites

 * Have your AWS Access Keys ready
 * If you want to use [CloudFlare](https://www.cloudflare.com), have your CloudFlare Token ready
 * For building you need to have [`builder`](https://github.com/giantswarm/builder) installed
 * When running Kocho in a Docker container, Docker needs to be installed

## Building Kocho

Building Kocho is rather easy. Just clone and make, and you are ready to go.

```
$ git clone https://github.com/giantswarm/kocho.git
$ cd kocho
$ make
```

The Makefile builds kocho by default for Linux. If you are running on Mac OS X,
please set `GOOS=darwin` before calling `make`:

```
$ GOOS=darwin make
```

You now should find a `kocho` binary file in your current folder.

## Configuring Kocho

First, let's initialize cloudconfig and cloudformation templates. By default
Kocho assumes they are in `templates/`.

```
kocho template-init
```

To actually use Kocho there needs to be a `kocho.yml` config file.

```
cluster_size: 3
certificate: <certificate>

machine_type: t2.micro

aws-vpc: <vpc>
aws-keypair: <keypair>
aws-subnet: <subnet>
aws-az: <availability zone>
```

To use CloudFlare for creating DNS records also add:

```
dns-sevice: cloudflare
dns-zone: <cloudflare domain>
```

To make Slack notifications work, put the slack configuration into `~/.giantswarm/kocho/slack.conf`.
```
{"token": "<slack token>", "username": "<slack username>", "notofication_channel": "<slack notification channel>"}
```

Further, make sure you have your AWS credentials in your environment.

```
export AWS_SECRET_ACCESS_KEY=<aws secret access key>
export AWS_ACCESS_KEY=<aws access key>
```

If you have configured Kocho in the `kocho.yml` to use CloudFlare. You also need to put your CloudFlare credentials into the environment.

```
export CLOUDFLARE_EMAIL=<cloudflare email>
export CLOUDFLARE_TOKEN=<cloudflare api token>
```

## Using Kocho

### Creating a cluster

Now we are going to create a new cluster called `test-getting-started`.

```
kocho create test-getting-started
```

### Listing Clusters

Once we created a cluster, we can check what we have using the `list` command.

```
kocho list
```

We should see something along these lines of:

```
Name                  Type        Created
test-getting-started  standalone  09 Feb 16 18:42 UTC
```

Now you are ready to use your AWS cluster. Once you no longer need it, it can be destroyed.

```
kocho destroy test-getting-started
```

By default you need to confirm the deletion.

```
are you sure you want to destroy 'test-getting-started'? Enter yes: yes
```
