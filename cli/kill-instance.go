package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/notification"
	"github.com/giantswarm/kocho/ssh"
	"github.com/giantswarm/kocho/swarm"
	"github.com/giantswarm/kocho/swarm/types"

	"github.com/juju/errgo"
)

var cmdKillInstance = &Command{
	Name:        "kill-instance",
	Description: "Kill one instance of a swarm",
	Summary:     "Kill and remove a swarm's instance from the etcd cluster",
	Run:         runKillInstance,
}

var (
	disableEtcdCleanup bool
)

func init() {
	cmdKillInstance.Flags.BoolVar(&disableEtcdCleanup, "disable-machine-cleanup", false, "do not connect to the machine and try to cleanly remove it from the etcd cluster first")
}

func runKillInstance(args []string) (exit int) {
	if len(args) != 2 {
		return exitError("wrong number of arguments. Usage: kocho kill-instance <swarm> <instance>")
	}
	swarmName := args[0]
	instanceID := args[1]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get instances of swarm: %s", swarmName), err)
	}

	instances, err := s.GetInstances()
	if err != nil {
		return exitError(err)
	}

	killableInstance, err := swarmtypes.FindInstanceById(instances, instanceID)
	if err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to find provided instance: %s", instanceID))
	}

	runningInstances := swarmtypes.FilterInstanceById(instances, instanceID)
	if len(runningInstances) == 0 {
		return exitError(errgo.Newf("no more instances left in swarm %s. Cannot update Fleet DNS entry", swarmName))
	}

	if !disableEtcdCleanup {

		if err = ssh.RemoveFromEtcd(killableInstance.PublicIPAddress); err != nil {
			return exitError(errgo.WithCausef(err, nil, "failed to remove instance from etcd cluster: %s", instanceID))
		}

		if err = ssh.StopEtcd(killableInstance.PublicIPAddress); err != nil {
			return exitError(errgo.WithCausef(err, nil, "failed to stop etcd on instance: %s", instanceID))
		}
	}

	if err = swarm.RemoveInstanceFromDiscovery(killableInstance); err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to remove instance from etcd discovery: %s", instanceID))
	}

	if err = s.KillInstance(killableInstance); err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to kill instance: %s", instanceID))
	}

	if changed, err := dns.Update(viperConfig.getDNSNamingPattern(), s, runningInstances); err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to update dns records"))
	} else if !changed {
		return exitError(errgo.Newf("DNS not changed. Couldn't find valid publid DNS name"))
	}

	notification.SendMessage(projectVersion, projectBuild)

	return 0
}
