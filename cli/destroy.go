package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/notification"
	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

var (
	cmdDestroy = &Command{
		Name:        "destroy",
		Description: "Destroy a swarm",
		Summary:     "Destroy a swarm on AWS",
		Run:         runDestroy,
	}

	forceDestroying = false
)

func init() {
	cmdDestroy.Flags.BoolVar(&sharedFlags.NoBlock, "no-block", false, "do not wait until the swarm has been deleted before exiting")
	cmdDestroy.Flags.BoolVar(&forceDestroying, "force", false, "do not confirm destroying")
}

func runDestroy(args []string) (exit int) {
	if len(args) == 0 {
		return exitError("no Swarm given. Usage: kocho destroy <swarm>")
	} else if len(args) > 1 {
		return exitError("too many arguments. Usage: kocho destroy <swarm>")
	}
	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't find swarm: %s", swarmName), err)
	}

	if !forceDestroying {
		if err := confirm(fmt.Sprintf("are you sure you want to destroy '%s'? Enter yes:", swarmName)); err != nil {
			return exitError("failed to read from stdin", err)
		}
	}

	if err := s.Destroy(); err != nil {
		return exitError(fmt.Sprintf("couldn't delete swarm: %s", swarmName), err)
	}

	err = dns.DeleteEntries(viperConfig.getDNSNamingPattern(), swarmName)
	if err != nil {
		return exitError("couldn't delete dns entries", err)
	}

	if !sharedFlags.NoBlock {
		err := s.WaitUntil(provider.StatusDeleted)
		if err != nil {
			return exitError("couldn't find out if swarm was deleted correctly", err)
		}
	} else {
		fmt.Printf("triggered swarm %s deletion\n", swarmName)
	}
	notification.SendMessage(projectVersion, projectBuild)

	return 0
}

func confirm(question string) error {
	for {
		fmt.Printf("%s ", question)
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			return err
		}

		if string(line) == "yes" {
			return nil
		}
		fmt.Println("please enter 'yes' to confirm")
	}
}
