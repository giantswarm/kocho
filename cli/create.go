package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/pflag"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/notification"
	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

var (
	cmdCreate = &Command{
		Name:        "create",
		Usage:       "<name>",
		Description: "Create a swarm",
		Summary:     "Create a new swarm on AWS",
		Run:         runCreate,
	}

	createShowCreateFlags bool
)

func init() {
	registerCreateFlags(&cmdCreate.Flags)

	cmdCreate.Flags.BoolVar(&createShowCreateFlags, "show-flags", false, "Prints the used parameters and quits.")
}

func registerCreateFlags(flagset *pflag.FlagSet) {
	flagset.String("type", "standalone", "type of the stack - there are primary, secondary and standalone stacks that form a cluster")
	flagset.String("tags", "", "tags that should be added to fleetd of the swarm (eg --tags=cluster=core,disk=ssd)")
	flagset.String("yochu", "", "version of Yochu to provision cluster nodes")
	flagset.Int("cluster-size", 3, "number of nodes a cluster should have")
	flagset.String("etcd-peers", "", "etcd peers a secondary swarm is connecting to")
	flagset.String("etcd-discovery-url", "", "etcd discovery url for a secondary swarm is connecting to")
	flagset.String("fleet-version", "v0.11.3-gs-2", "version to use when provisioning fleetd/fleetctl binaries")
	flagset.String("etcd-version", "v2.1.2-gs-1", "version to use when provisioning etcd/etcdctl binaries")
	flagset.String("template-dir", "templates", "directory to use for reading templates (see template-init command)")
	flagset.String("docker-version", "1.6.2", "version to use when provisioning docker binaries")
	flagset.String("template-dir", "templates", "directory to use for reading templates")

	awsEuWest1CoreOS := "ami-5f2f5528" // CoreOS Stable 681.2.0 (HVM eu-west-1)
	flagset.String("image", awsEuWest1CoreOS, "image version that should be used to create a swarm")
	flagset.String("certificate", "", "certificate ARN to use to create aws cluster")
	flagset.String("machine-type", "m3.large", "machine type to use, e.g. m3.large for AWS")

	// AWS Provider specific
	flagset.String("aws-keypair", "", "Keypair to use for AWS machines")
	flagset.String("aws-vpc", "", "VPC to use for new AWS machines")
	flagset.String("aws-vpc-cidr", "", "VPC CIDR to use for security configuration")
	flagset.String("aws-subnet", "", "Subnet to use for new AWS machines")
	flagset.String("aws-az", "", "AZ to use for new AWS machines")
}

func runCreate(args []string) (exit int) {
	flags := viperConfig.newViperCreateFlags()

	if createShowCreateFlags {
		data, err := json.MarshalIndent(flags, "", "  ")
		if err != nil {
			exitError("Failed to json encode flags: %v", err)
		}
		fmt.Printf("%s\n", string(data))
		return
	}

	if flags.FleetVersion == "" {
		return exitError("couldn't create swarm: fleet version must be set using --fleet-version=<version>")
	}

	if flags.EtcdVersion == "" {
		return exitError("couldn't create swarm: etcd version must be set using --etcd-version=<version>")
	}

	if flags.MachineType == "" {
		return exitError("couldn't create swarm: --machine-type must be provided")
	}
	if flags.ImageURI == "" {
		return exitError("couldn't create swarm: --image must be provided")
	}

	if len(args) == 0 {
		return exitError("no Swarm given. Usage: kocho create <swarm>")
	} else if len(args) > 1 {
		return exitError("too many arguments. Usage: kocho create <swarm>")
	}
	name := args[0]

	s, err := swarmService.Create(name, swarm.AWS, flags)
	if err != nil {
		return exitError(fmt.Sprintf("couldn't create swarm: %s", name), err)
	}

	if !sharedFlags.NoBlock {
		err = s.WaitUntil(provider.StatusCreated)
		if err != nil {
			return exitError("couldn't find out if swarm was started correctly", err)
		}

		err = dns.CreateSwarmEntries(viperConfig.getDNSNamingPattern(), s)
		if err != nil {
			return exitError("couldn't create dns entries", err)
		}
	} else {
		fmt.Printf("triggered swarm %s start. No DNS will be configured\n", name)
	}
	notification.SendMessage(projectVersion, projectBuild)

	return 0
}