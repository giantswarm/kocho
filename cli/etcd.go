package cli

import (
	"fmt"
	"strings"

	"github.com/giantswarm/kocho/ssh"
	"github.com/giantswarm/kocho/swarm"

	"github.com/juju/errgo"
)

var cmdEtcd = &Command{
	Name:        "etcd",
	Description: "Get the etcd details of a swarm",
	Summary:     "Get the etcd details of a swarm",
	Run:         runEtcd,
}

func runEtcd(args []string) (exit int) {
	if len(args) != 2 {
		return exitError("usage: kocho etcd discovery|peers <swarm>")
	}

	subCommand := args[0]
	swarmName := args[1]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get instances of swarm: %s", swarmName), err)
	}

	instances, err := s.GetInstances()
	if err != nil {
		exitError(err)
	}

	if len(instances) == 0 {
		return exitError(errgo.Newf("could not find any running instances in swarm %s", swarmName))
	}

	switch subCommand {
	case "discovery":
		url, err := ssh.GetEtcdDiscoveryUrl(instances[0].PublicIPAddress)
		if err != nil {
			return exitError(err)
		}

		fmt.Println(url)
	case "peers":
		etcdPeers := []string{}
		for _, instance := range instances {
			etcdPeers = append(etcdPeers, fmt.Sprintf("http://%v:2379", instance.PrivateIPAddress))
		}
		peers := strings.Join(etcdPeers, ",")

		fmt.Println(peers)
	}

	return 0
}
