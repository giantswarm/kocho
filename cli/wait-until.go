package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

var waitUntilCmd = &cobra.Command{
	Use:   "wait-until [swarm_name] [desired_state]",
	Short: "Wait until a swarm has reached a certain status",
	Long:  "Wait until a swarm has reached a certain status, such as the swarm being created or deleted",
	Run:   runWaitUntil,
}

func init() {
	RootCmd.AddCommand(waitUntilCmd)
}

func runWaitUntil(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Usage()
		return
	}

	name := args[0]
	status := args[1]

	s, err := swarmService.Get(name, swarm.AWS)
	if err != nil {
		if status == "deleted" && err == provider.ErrNotFound {
			return
		} else {
			fmt.Printf("couldn't find swarm: %s\n", err)
			return
		}
	}

	err = s.WaitUntil(status)
	if err != nil {
		fmt.Printf("swarm didn't reach desired state: %s\n", err)
	}
}
