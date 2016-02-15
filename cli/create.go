package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm"
)

const (
	awsEuWest1CoreOS = "ami-5f2f5528" // CoreOS Stable 681.2.0 (HVM eu-west-1)
)

var (
	createCmd = &cobra.Command{
		Use:   "create [swarm_name]",
		Short: "Create a swarm",
		Long:  "Create a swarm on AWS, with given configuration",
		Run:   runCreate,
	}

	createShowCreateFlags bool
)

func init() {
	createCmd.Flags().String("type", "standalone", "type of the stack - there are primary, secondary and standalone stacks that form a cluster")
	createCmd.Flags().String("tags", "", "tags that should be added to fleetd of the swarm (eg --tags=cluster=core,disk=ssd)")
	createCmd.Flags().Int("cluster-size", 3, "number of nodes a cluster should have")
	createCmd.Flags().String("etcd-peers", "", "etcd peers for a secondary swarm to connect to")
	createCmd.Flags().String("etcd-discovery-url", "", "etcd discovery url for a secondary swarm to connect to")
	createCmd.Flags().String("template-dir", "templates", "directory to use for reading templates (see template-init command)")

	createCmd.Flags().String("image", awsEuWest1CoreOS, "image version that should be used to create a swarm")
	createCmd.Flags().String("certificate", "", "certificate ARN to use to create aws cluster")
	createCmd.Flags().String("machine-type", "m3.large", "machine type to use (e.g. m3.large for AWS)")

	// Yochu
	createCmd.Flags().String("yochu", "", "version of Yochu to use when provisioning cluster nodes")
	createCmd.Flags().String("yochu-docker-version", "1.6.2", "version to use when provisioning docker binaries")
	createCmd.Flags().String("yochu-fleet-version", "v0.11.3-gs-2", "version to use when provisioning fleetd/fleetctl binaries")
	createCmd.Flags().String("yochu-etcd-version", "v2.1.2-gs-1", "version to use when provisioning etcd/etcdctl binaries")

	// AWS Provider specific
	createCmd.Flags().String("aws-keypair", "", "keypair to use for AWS machines")
	createCmd.Flags().String("aws-vpc", "", "VPC to use for new AWS machines")
	createCmd.Flags().String("aws-vpc-cidr", "", "VPC CIDR to use for security configuration")
	createCmd.Flags().String("aws-subnet", "", "subnet to use for new AWS machines")
	createCmd.Flags().String("aws-az", "", "AZ to use for new AWS machines")

	createCmd.Flags().BoolVar(&createShowCreateFlags, "show-flags", false, "print the used parameters and quit")

	RootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) {
	flags := viperConfig.newViperCreateFlags()

	if createShowCreateFlags {
		data, err := json.MarshalIndent(flags, "", "  ")
		if err != nil {
			fmt.Printf("Failed to json encode flags: %s\n", err)
			return
		}

		fmt.Printf("%s\n", string(data))
		return
	}

	if len(args) != 1 {
		cmd.Usage()
		return
	}

	if flags.FleetVersion == "" {
		fmt.Println("couldn't create swarm: fleet version must be set using --fleet-version=<version>")
		return
	}

	if flags.EtcdVersion == "" {
		fmt.Println("couldn't create swarm: etcd version must be set using --etcd-version=<version>")
		return
	}

	if flags.MachineType == "" {
		fmt.Println("couldn't create swarm: --machine-type must be provided")
		return
	}

	if flags.ImageURI == "" {
		fmt.Println("couldn't create swarm: --image must be provided")
		return
	}

	name := args[0]

	s, err := swarmService.Create(name, swarm.AWS, flags)
	if err != nil {
		fmt.Printf("couldn't create swarm: %s\n", err)
		return
	}

	if !sharedFlags.NoBlock {
		err = s.WaitUntil(provider.StatusCreated)
		if err != nil {
			fmt.Printf("couldn't find out if swarm was started correctly: %s\n", err)
			return
		}

		err = dns.CreateSwarmEntries(dnsService, viperConfig.getDNSNamingPattern(), s)
		if err != nil {
			fmt.Printf("couldn't create dns entries: %s\n", err)
			return
		}
	} else {
		fmt.Printf("triggered swarm %s start. No DNS will be configured\n", name)
	}

	fireNotification()
}
