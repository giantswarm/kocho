package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

var (
	destroyCmd = &cobra.Command{
		Use:     "destroy [swarm_name]",
		Aliases: []string{"rm"},
		Short:   "Destroy a swarm",
		Long:    "Obliterate a swarm from existence",
		Run:     runDestroy,
	}

	forceDestroying = false
)

func init() {
	destroyCmd.Flags().BoolVar(&sharedFlags.NoBlock, "no-block", false, "do not wait until the swarm has been deleted before exiting")
	destroyCmd.Flags().BoolVar(&forceDestroying, "force", false, "do not confirm before destroying")

	RootCmd.AddCommand(destroyCmd)
}

func runDestroy(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't find swarm: %s\n", err)
		return
	}

	if !forceDestroying {
		if err := confirm(fmt.Sprintf("are you sure you want to destroy '%s'? Enter yes:", swarmName)); err != nil {
			fmt.Printf("failed to read from stdin: %s\n", err)
			return
		}
	}

	if err := s.Destroy(); err != nil {
		fmt.Printf("couldn't delete swarm: %s\n", err)
		return
	}

	err = dns.DeleteEntries(dnsService, viperConfig.getDNSNamingPattern(), swarmName)
	if err != nil {
		fmt.Printf("couldn't delete DNS entries: %s\n", err)
		return
	}

	if !sharedFlags.NoBlock {
		err := s.WaitUntil(provider.StatusDeleted)
		if err != nil {
			fmt.Printf("couldn't find out if swarm was deleted correctly: %s\n", err)
			return
		}
	} else {
		fmt.Printf("triggered swarm %s deletion\n", swarmName)
	}

	fireNotification()
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
