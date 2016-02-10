package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/swarm"
)

var cmdStatus = &Command{
	Name:        "status",
	Description: "Status of a swarm",
	Summary:     "Status of a swarm",
	Run:         runStatus,
}

func runStatus(args []string) (exit int) {
	if len(args) == 0 {
		return exitError("no Swarm given. Usage: kocho status <swarm>")
	} else if len(args) > 1 {
		return exitError("too many arguments. Usage: kocho status <swarm>")
	}
	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get status of swarm: %s", swarmName), err)
	}

	status, reason, err := s.GetStatus()
	if err != nil {
		return exitError(fmt.Sprintf("couldn't get status of swarm: %s", swarmName), err)
	}
	fmt.Println(status, reason)
	return 0
}
