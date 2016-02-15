package cli

import (
	"fmt"
	"strings"

	"github.com/giantswarm/kocho/ssh"
	"github.com/giantswarm/kocho/swarm"

	"github.com/spf13/cobra"
)

var (
	etcdCmd = &cobra.Command{
		Use:   "etcd",
		Short: "Manage etcd of a swarm",
		Long:  "Manage etcd discovery URLs and peers of a swarm",
	}

	etcdDiscoveryCmd = &cobra.Command{
		Use:   "discovery [swarm_name]",
		Short: "Get etcd discovery URL",
		Long:  "Print out the etcd discovery URL used to bootstrap a swarm",
		Run:   runEtcdDiscoveryCmd,
	}

	etcdPeersCmd = &cobra.Command{
		Use:   "peers [swarm_name]",
		Short: "Get etcd peers",
		Long:  "Print out the peers of the etcd cluster",
		Run:   runEtcdPeersCmd,
	}
)

func init() {
	RootCmd.AddCommand(etcdCmd)

	etcdCmd.AddCommand(etcdDiscoveryCmd)
	etcdCmd.AddCommand(etcdPeersCmd)
}

func runEtcdDiscoveryCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	instances, err := s.GetInstances()
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	if len(instances) == 0 {
		fmt.Printf("could not find any running instances in swarm\n")
		return
	}

	url, err := ssh.GetEtcdDiscoveryUrl(instances[0].PublicIPAddress)
	if err != nil {
		fmt.Printf("could not get etcd discovery url: %s\n", err)
		return
	}

	fmt.Println(url)
}

func runEtcdPeersCmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	instances, err := s.GetInstances()
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	if len(instances) == 0 {
		fmt.Printf("could not find any running instances in swarm\n")
		return
	}

	etcdPeers := []string{}
	for _, instance := range instances {
		etcdPeers = append(etcdPeers, fmt.Sprintf("http://%v:2379", instance.PrivateIPAddress))
	}
	peers := strings.Join(etcdPeers, ",")

	fmt.Println(peers)
}
