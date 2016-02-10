package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/swarm"
	"github.com/ryanuber/columnize"
)

var cmdInstances = &Command{
	Name:        "instances",
	Description: "List instances of a swarm",
	Summary:     "List all the dns names of the instances of a swarm",
	Run:         runInstances,
}

const (
	instancesHeader = "Id | Image | Type | PublicDns | PrivateDns"
	instancesScheme = "%s | %s | %s | %s | %s"
)

func runInstances(args []string) (exit int) {
	if len(args) == 0 {
		return exitError("no Swarm given. Usage: kocho instances <swarm>")
	} else if len(args) > 1 {
		return exitError("too many arguments. Usage: kocho instances <swarm>")
	}
	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get instances of swarm: %s", swarmName), err)
	}

	instances, err := s.GetInstances()
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get instances of swarm: %s", swarmName), err)
	}

	lines := []string{instancesHeader}
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf(instancesScheme, i.Id, i.Image, i.Type, i.PublicDNSName, i.PrivateDNSName))
	}
	fmt.Println(columnize.SimpleFormat(lines))
	return 0
}
