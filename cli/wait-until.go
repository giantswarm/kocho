package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

var (
	cmdWaitUntil = &Command{
		Name:        "wait-until",
		Summary:     "Wait until a swarm has a certain status",
		Usage:       "{SWARM} {DESIRED_STATE}",
		Description: "This is useful to wait until a swarm has been created or deleted.",
		Run:         runWaitUntil,
	}
)

func runWaitUntil(args []string) (exit int) {
	if len(args) < 2 {
		return exitError("wrong amount of arguments. Usage: kocho wait-until <swarm> <status>")
	} else if len(args) > 2 {
		return exitError("too many arguments. Usage: kocho wait-until <swarm> <status>")
	}

	name := args[0]
	status := args[1]

	s, err := swarmService.Get(name, swarm.AWS)
	if err != nil {
		if status == "deleted" && err == provider.ErrNotFound {
			return 0
		} else {
			return exitError(fmt.Sprintf("couldn't find swarm: %s", name), err)
		}
	}

	err = s.WaitUntil(status)
	if err != nil {
		return exitError(fmt.Sprintf("swarm didn't reach desired state: %s", status), err)
	}

	return 0
}
