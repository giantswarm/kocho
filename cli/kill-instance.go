package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/ssh"
	"github.com/giantswarm/kocho/swarm"
	"github.com/giantswarm/kocho/swarm/types"
)

var (
	killInstanceCmd = &cobra.Command{
		Use:   "kill-instance [swarm_name] [instance_id]",
		Short: "Kill an instance of a swarm",
		Long:  "Kill and remove an instance from the etcd cluster",
		Run:   runKillInstance,
	}

	disableEtcdCleanup bool
)

func init() {
	killInstanceCmd.Flags().BoolVar(&disableEtcdCleanup, "disable-machine-cleanup", false, "do not try to remove the machine from the etcd cluster before killing it")

	RootCmd.AddCommand(killInstanceCmd)
}

func runKillInstance(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Usage()
		return
	}

	swarmName := args[0]
	instanceID := args[1]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't get swarm: %s\n", err)
		return
	}

	instances, err := s.GetInstances()
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	killableInstance, err := swarmtypes.FindInstanceById(instances, instanceID)
	if err != nil {
		fmt.Printf("failed to find provided instance: %s\n", err)
		return
	}

	runningInstances := swarmtypes.FilterInstanceById(instances, instanceID)
	if len(runningInstances) == 0 {
		fmt.Printf("no instances in swarm - cannot update fleet DNS entry\n")
		return
	}

	if !disableEtcdCleanup {
		if err = ssh.RemoveFromEtcd(killableInstance.PublicIPAddress); err != nil {
			fmt.Printf("failed to remove instance from etcd cluster: %s\n", err)
			return
		}

		if err = ssh.StopEtcd(killableInstance.PublicIPAddress); err != nil {
			fmt.Printf("failed to stop etcd on instance: %s\n", err)
			return
		}
	}

	if err = swarm.RemoveInstanceFromDiscovery(killableInstance); err != nil {
		fmt.Printf("failed to remove instance from etcd discovery: %s\n", err)
		return
	}

	if err = s.KillInstance(killableInstance); err != nil {
		fmt.Printf("failed to kill instance: %s\n", err)
		return
	}

	if changed, err := dns.Update(dnsService, viperConfig.getDNSNamingPattern(), s, runningInstances); err != nil {
		fmt.Printf("failed to update dns records: %s\n", err)
		return
	} else if !changed {
		fmt.Printf("DNS not changed. Couldn't find valid publid DNS name: %s\n", err)
		return
	}

	fireNotification()
}
