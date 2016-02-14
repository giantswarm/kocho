package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/swarm"
	"github.com/ryanuber/columnize"
)

const (
	instancesHeader = "Id | Image | Type | PublicDns | PrivateDns"
	instancesScheme = "%s | %s | %s | %s | %s"
)

var instancesCmd = &cobra.Command{
	Use:   "instances [swarm_name]",
	Short: "List instances of a swarm",
	Long:  "List instances of a swarm, with type and dns values",
	Run:   runInstances,
}

func init() {
	RootCmd.AddCommand(instancesCmd)
}

func runInstances(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	swarmName := args[0]

	s, err := swarmService.Get(swarmName, swarm.AWS)
	if err != nil {
		fmt.Printf("couldn't get swarm: %s\n", err)
		return
	}

	instances, err := s.GetInstances()
	if err != nil {
		fmt.Printf("couldn't get instances of swarm: %s\n", err)
		return
	}

	lines := []string{instancesHeader}
	for _, i := range instances {
		lines = append(lines, fmt.Sprintf(instancesScheme, i.Id, i.Image, i.Type, i.PublicDNSName, i.PrivateDNSName))
	}

	fmt.Println(columnize.SimpleFormat(lines))
}
