// Package cli provides commands to use Kocho.
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/provider/aws/sdk"
	"github.com/giantswarm/kocho/swarm"

	"github.com/spf13/pflag"
)

const (
	cliName        = "kocho"
	cliDescription = "kocho is a command-line interface to control swarms."
)

var (
	out           *tabwriter.Writer
	globalFlagset = pflag.NewFlagSet(cliName, pflag.ExitOnError)

	// top level commands
	commands []*Command

	// flags used by all commands
	globalFlags = struct {
		Debug   bool
		Version bool
		Quiet   bool
		Help    bool
	}{}

	// flags used by multiple commands
	sharedFlags = struct {
		NoBlock bool
	}{}

	// bumped project version. Will be overriden by the compiler
	projectVersion = "dev"
	projectBuild   string

	viperConfig *KochoConfiguration

	dnsService dns.DNSService

	swarmService      *swarm.Service
	swarmConfig       swarm.Config
	swarmDependencies swarm.Dependencies
)

func init() {
	globalFlagset.BoolVar(&globalFlags.Debug, "debug", false, "print out more debug information to stderr")
	globalFlagset.BoolVar(&globalFlags.Version, "version", false, "print the version and exit")
	globalFlagset.BoolVar(&globalFlags.Quiet, "quiet", false, "be quiet on output")
	globalFlagset.BoolVarP(&globalFlags.Help, "help", "h", false, "shows the help")

	// DNS Specific (used by create, kill-instance, dns subcmds)
	// see config.go getDNSNamingPattern()
	globalFlagset.String("dns-service", "cloudflare", "The DNS backend to use, defaults to cloudflare")
	globalFlagset.String("dns-zone", dns.DefaultNamingPattern.Zone, "the zone to create the dns records in")
	globalFlagset.String("dns-catchall", dns.DefaultNamingPattern.Catchall, "template for the catchall dns record")
	globalFlagset.String("dns-catchall-private", dns.DefaultNamingPattern.CatchallPrivate, "template for the catchall-private dns record")
	globalFlagset.String("dns-public", dns.DefaultNamingPattern.Public, "template for the public dns record")
	globalFlagset.String("dns-private", dns.DefaultNamingPattern.Private, "template for the private dns record")
	globalFlagset.String("dns-fleet", dns.DefaultNamingPattern.Fleet, "template for the fleet dns record")

	sdk.DefaultSessionProvider.RegisterFlagSet(globalFlagset)
}

// Command describes a command that can be run.
type Command struct {
	Name        string        // Name of the Command and the string to use to invoke it
	Summary     string        // One-sentence summary of what the Command does
	Usage       string        // Usage options/arguments
	Description string        // Detailed description of command
	Flags       pflag.FlagSet // Set of flags associated with this command

	Run func(args []string) int // Run a command with the given arguments, return exit status
}

// NewKochoCmd configures available commands, and runs them.
func NewKochoCmd() {
	out = new(tabwriter.Writer)
	out.Init(os.Stdout, 0, 8, 1, '\t', 0)
	commands = []*Command{
		cmdCreate,
		cmdDestroy,
		cmdInstances,
		cmdKillInstance,
		cmdEtcd,
		cmdStatus,
		cmdList,
		cmdWaitUntil,
		cmdDns,
		cmdHelp,
		cmdVersion,
		cmdTemplateInit,
		cmdSlack,
	}

	Execute()
}

// Execute loads configuration, determines the necessary command to run, and runs it.
func Execute() {
	viperConfig = NewConfig()
	if err := viperConfig.loadConfig(); err != nil {
		os.Exit(exitError("failed to read config file: ", err))
	}

	globalFlagset.SetInterspersed(false)
	globalFlagset.Parse(os.Args[1:])

	if err := viperConfig.configFromPFlags(globalFlagset); err != nil {
		fmt.Println("failed to copy arguments to viper")
		fmt.Println(err.Error())
		os.Exit(2)
	}

	var args = globalFlagset.Args()
	if len(args) < 1 || globalFlags.Help {
		args = append([]string{"help"}, args...)
	}

	// deal specially with --version
	if globalFlags.Version {
		args[0] = "version"
	}

	var cmd *Command

	// determine which Command should be run
	for _, c := range commands {
		if c.Name == args[0] {
			cmd = c
			break
		}
	}

	if cmd == nil {
		fmt.Printf("%v: unknown subcommand: %q\n", cliName, args[0])
		fmt.Printf("run '%v help' for usage\n", cliName)
		os.Exit(2)
	}

	// Init global stuff for the CLI, e.g. the swarm service
	dnsService = newDNSService(viperConfig)
	swarmService = swarm.NewService(swarmConfig, swarmDependencies)

	// Copy command specific flags into viper
	if err := viperConfig.configFromPFlags(&cmd.Flags); err != nil {
		fmt.Println("failed to copy arguments to viper")
		fmt.Println(err.Error())
		os.Exit(2)
	}

	if err := cmd.Flags.Parse(args[1:]); err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	os.Exit(cmd.Run(cmd.Flags.Args()))
}

func getAllFlags() (flags []*pflag.Flag) {
	return getFlags(globalFlagset)
}

func getFlags(flagset *pflag.FlagSet) (flags []*pflag.Flag) {
	flags = make([]*pflag.Flag, 0)
	flagset.VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f)
	})
	return
}

func exitError(args ...interface{}) (exit int) {
	fmt.Fprintln(os.Stderr, args...)
	return 1
}
