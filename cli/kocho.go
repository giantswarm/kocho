// Package cli provides commands to use Kocho.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/provider/aws/sdk"
	"github.com/giantswarm/kocho/swarm"
)

var RootCmd = &cobra.Command{
	Use:   "kocho",
	Short: "Kocho sets up CoreOS clusters",
	Long:  "Kocho sets up CoreOS clusters on AWS",
}

var (
	// flags used by all commands
	globalFlags = struct {
		Debug bool
		Quiet bool
	}{}

	// flags used by multiple commands
	sharedFlags = struct {
		NoBlock bool
	}{}

	// bumped project version. Will be overriden by the compiler
	projectVersion = "dev"
	projectBuild   string

	viperConfig *KochoConfiguration

	swarmService      *swarm.Service
	swarmConfig       swarm.Config
	swarmDependencies swarm.Dependencies

	dnsService dns.DNSService
)

func init() {
	RootCmd.PersistentFlags().BoolVar(&globalFlags.Debug, "debug", false, "print out debug information to stderr")
	RootCmd.PersistentFlags().BoolVar(&globalFlags.Quiet, "quiet", false, "be quiet on ouput")

	// DNS Specific (used by create, kill-instance, dns subcmds)
	// see config.go getDNSNamingPattern()
	dnsFlags := []struct {
		name  string
		value string
		usage string
	}{
		{"dns-service", "cloudflare", "DNS backend to use"},
		{"dns-zone", dns.DefaultNamingPattern.Zone, "zone to create the dns records in"},
		{"dns-catchall", dns.DefaultNamingPattern.Catchall, "template for the catchall dns record"},
		{"dns-catchall-private", dns.DefaultNamingPattern.CatchallPrivate, "template for the catchall-private dns record"},
		{"dns-public", dns.DefaultNamingPattern.Public, "template for the public dns record"},
		{"dns-private", dns.DefaultNamingPattern.Private, "template for the private dns record"},
		{"dns-fleet", dns.DefaultNamingPattern.Fleet, "template for the fleet dns record"},
	}
	dnsCommands := []*cobra.Command{createCmd, dnsCmd, killInstanceCmd}

	for _, cmd := range dnsCommands {
		for _, dnsFlag := range dnsFlags {
			cmd.Flags().String(dnsFlag.name, dnsFlag.value, dnsFlag.usage)
		}
	}

	sdk.DefaultSessionProvider.RegisterFlagSet(RootCmd.Flags())

	viperConfig = NewConfig()
	if err := viperConfig.loadConfig(); err != nil {
		fmt.Println("failed to read config file: %s\n", err)
		return
	}

	// Copy all flags into viper
	commands := RootCmd.Commands()
	commands = append(commands, RootCmd)
	for _, command := range commands {
		if err := viperConfig.configFromPFlags(command.Flags()); err != nil {
			fmt.Println("failed to copy arguments to viper: %s", command.Name())
			fmt.Println(err.Error())
			os.Exit(2)
		}
	}

	// Init global stuff for the CLI, e.g. the swarm service
	dnsService = newDNSService(viperConfig)
	swarmService = swarm.NewService(swarmConfig, swarmDependencies)
}
