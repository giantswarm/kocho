package cli

import (
	"fmt"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/notification"
	"github.com/giantswarm/kocho/swarm"
)

var (
	flagDelete bool
	cmdDns     = &Command{
		Name:        "dns",
		Summary:     "Update DNS of a swarm",
		Usage:       "[--delete] <swarm>",
		Description: "Updating public, private and fleet dns entries of a swarm. With the --delete flag you can also just delete the current DNS entries",
		Run:         runDns,
	}
)

func init() {
	cmdDns.Flags.BoolVar(&flagDelete, "delete", false, "delete DNS entries of a swarm")
}

func runDns(args []string) (exit int) {
	if len(args) == 0 {
		return exitError("no Swarm given. Usage: kocho dns <swarm>")
	} else if len(args) > 1 {
		return exitError("too many arguments. Usage: kocho dns <swarm>")
	}
	name := args[0]

	err := dns.DeleteEntries(viperConfig.getDNSNamingPattern(), name)
	if err != nil {
		return exitError("couldn't delete dns entries", err)
	}

	if !flagDelete {
		s, err := swarmService.Get(name, swarm.AWS)
		if err != nil {
			return exitError(fmt.Sprintf("couldn't find swarm: %s", name), err)
		}

		err = dns.CreateSwarmEntries(viperConfig.getDNSNamingPattern(), s)
		if err != nil {
			return exitError("couldn't update dns entries", err)
		}
	}
	notification.SendMessage(projectVersion, projectBuild)

	return 0
}
