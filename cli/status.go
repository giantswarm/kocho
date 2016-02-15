package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/swarm"
)

var statusCmd = &cobra.Command{
	Use:   "status [swarm_name]",
	Short: "Status of a swarm",
	Long:  "Prints out the status of a given swarm",
	Run:   runStatus,
}

func init() {
	RootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't get status of swarm: %s\n", err)
		return
	}

	status, reason, err := s.GetStatus()
	if err != nil {
		fmt.Printf("couldn't get status of swarm: %s\n", swarmName)
		return
	}

	fmt.Println(status, reason)
}
