package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/swarm"
)

var (
	dnsCmd = &cobra.Command{
		Use:   "dns [swarm_name]",
		Short: "Update DNS entries of a swarm",
		Long:  "Update public, private and fleet DNS entries of a swarm",
		Run:   runDns,
	}

	flagDelete bool
)

func init() {
	dnsCmd.Flags().BoolVar(&flagDelete, "delete", false, "delete DNS entries of a swarm")

	RootCmd.AddCommand(dnsCmd)
}

func runDns(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	name := args[0]

	err := dns.DeleteEntries(dnsService, viperConfig.getDNSNamingPattern(), name)
	if err != nil {
		fmt.Printf("couldn't delete DNS entries: %s\n", err)
		return
	}

	if !flagDelete {
		s, err := swarmService.Get(name, swarm.AWS)
		if err != nil {
			fmt.Printf("couldn't find swarm: %s\n", err)
			return
		}

		err = dns.CreateSwarmEntries(dnsService, viperConfig.getDNSNamingPattern(), s)
		if err != nil {
			fmt.Printf("couldn't update DNS entries: %s\n", err)
			return
		}
	}

	fireNotification()
}
