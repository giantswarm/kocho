package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/dns"
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

const (
	etcdDocsLink = "http://github.com/giantswarm/kocho/blob/master/docs/etcd-operations.md"

	// Params: instanceId, etcdDocsLink
	killInstanceSuccessMessage = "Success! Instance %s has been terminated.\n" +
		"The Autoscaler will start a new instance in its place.\n" +
		"If you want the new instance to be part of the etcd quorum, please add it, when it's up. \n" +
		"Follow the guide at %s for help.\n"
)

var (
	ignoreQuorumCheck bool
)

func init() {
	cmdKillInstance.Flags.BoolVar(&ignoreQuorumCheck, "ignore-quorum-check", false, "do not connect to the machine and check if it is part of the etcd quorum")
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

	if !ignoreQuorumCheck {
		etcdQuorumID, err := ssh.GetEtcd2MemberName(killableInstance.PublicIPAddress)
		if err != nil {
			return exitError(errgo.WithCausef(err, nil, "ssh: failed to check quorum member list: %v", err))
		}

		if etcdQuorumID != "" {
			return exitError(errgo.Newf("Instance %s seems to be part of the etcd quorum. Please remove it beforehand. See %s", killableInstance.Id, etcdDocsLink))
		}
	}

	if err = s.KillInstance(killableInstance); err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to kill instance: %s", instanceID))
	}

	if changed, err := dns.Update(dnsService, viperConfig.getDNSNamingPattern(), s, runningInstances); err != nil {
		return exitError(errgo.WithCausef(err, nil, "failed to update dns records"))
	} else if !changed {
		return exitError(errgo.Newf("DNS not changed. Couldn't find valid publid DNS name"))
	}

	fmt.Printf(killInstanceSuccessMessage, killableInstance.Id, etcdDocsLink)

	fireNotification()

	return 0
}
