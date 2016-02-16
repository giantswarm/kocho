# Etcd Operations

The following guides are for etcd2 and should be performed with care. Make sure you have backups of the datadirs beforehand.

## Removing a member from the quorum

To remove a member from the quorum you need its ID. Use `etcdctl member list` on one of the machines to get it:

```
$ etcdctl member list | grep "name=$(cat /etc/machine-id)" | cut -f1 -d:
abcdef123456
```

Now, you need to remove it from the quorum itself:

```
$ etcdctl member remove abcdef123456
```

It's also advised to remove it from the etcd discovery document. You can use `kocho etcd discovery <cluster>` on your machine for this.

```
$ kocho etcd discovery <clustername>
https://discovery.etcd.io/544503849b0bbb16263321f824e6643f

# Replace the last part here with the machine's ID from before
$ curl -XDELETE https://discovery.etcd.io/544503849b0bbb16263321f824e6643f/abcdef123456
```

## Promoting a machine to be part of the quorum

TBD
This operation is quite **IMPORTANT** and **CRITICAL**, so if you don't feel READY to do it, or your knowledge of etcd is not sufficient to face sudden errors. Please, ASK one of your colleagues, you can _BRING DOWN_ the whole cluster.

1. You need to verify if the member(machine) you want to move out, it's an etcd leader or not. If this machine is an etcd leader, you should be CAREFUL and force a leader election prior to continue the member removal operation.
2. Once your etcd node is not the leader one, you can remove as etcd member.
    * `etcdctl member remove ${ID_etcdctl_member_list}`
3. Remove this node from the discovery
    * `cat etcd2.service.d/20-cloudinit.conf` to get the DISCOVERY_ID
    * curl https://discovery.etcd.io/DISCOVERY_ID/{ETCDCTL_MEMBER_ID} -XDELETE

4. Add new member to the core members:

```
export ETCD_DATA_DIR=/var/lib/etcd2
export ETCD_NAME=$(cat /etc/machine-id)
export ETCD_INITIAL_CLUSTER_STATE=existing

source /etc/environment
export private_ipv4=$COREOS_PRIVATE_IPV4

export ETCD_ADVERTISE_CLIENT_URLS=http://$private_ipv4:2379;
export ETCD_INITIAL_ADVERTISE_PEER_URLS=http://$private_ipv4:2380;
export ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:2379;
export ETCD_LISTEN_PEER_URLS=http://$private_ipv4:2380;

# start etcd2 

details=$(etcdctl member add $ETCD_NAME http://$private_ipv4:2380); 
eval export `echo "$details" | tail -n-3`;                          


$ sudo systemctl stop etcd2;                                               


# stop etcd2 
$ sudo rm -rf /var/lib/etcd2/proxy
#etcd2 has to be started once with the current `env`, wait for few seconds to verify the execution is correct and there isn't errors.
$ sudo -E -u etcd /home/core/bin/etcd2

$ pkill etcd2

$ sudo systemctl restart etcd2
```

5. Add your node to the discovery with the following command:
`curl --data-urlencode "value=$MACHINE_ID=http://$MACHINE_IP:2380" -XPUT https://discovery.etcd.io/$CLUSTER_ID/$ID_ETCD_MEMBER_LIST`

### List of IMPORTANT Checks prior to close your Terminal and go for a beer
1. Check if the node has been added `etcdctl member list`. Verify that all the data is correct. It can happen that for some reason some fields are not fulfill, if so ASK to one of your colleagues.
2. Verify if the cluster is healthy `etcdctl cluster-health`. It has to be, if not you should ASK to one of your colleagues for help.
3. Get the discovery information to figure it out if your new member is included as an etcd node in the discovery information, `curl https://discovery.etcd.io/DISCOVERY_ID`. It has to be there, if not you should ASK to one of your colleagues for help.