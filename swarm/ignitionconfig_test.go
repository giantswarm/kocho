package swarm

import (
	"strings"
	"testing"

	"github.com/giantswarm/kocho/swarm/types"
)

// getDefaultTestIgnitionCreateFlags returns a CreateFlags that can be used for testing
func getDefaultTestIgnitionCreateFlags(templateType string) swarmtypes.CreateFlags {
	awsCreateFlags := swarmtypes.AWSCreateFlags{
		KeypairName:      "keypair",
		VPC:              "vpc-8ad7213f",
		Subnet:           "subnet-5736g493",
		AvailabilityZone: "eu-west-1a",
	}

	flags := swarmtypes.CreateFlags{
		Type:             templateType,
		Tags:             "",
		YochuVersion:     "0.9.0",
		ClusterSize:      3,
		EtcdPeers:        "",
		FleetVersion:     "v0.11.3-gs-2",
		EtcdVersion:      "v2.1.2-gs-1",
		RktVersion:       "v1.1.0",
		K8sVersion:       "v1.2.0",
		DockerVersion:    "v1.9.1",
		EtcdDiscoveryURL: "",
		TemplateDir:      "../default-templates", // Hack to use default built in templates
		MachineType:      "m3.large",
		ImageURI:         "ami-5k2l4639",
		CertificateURI:   "arn:aws:iam::063170442734:server-certificate/wildcard-2015-chained",
		AWSCreateFlags:   &awsCreateFlags,

		UseIgnition: true,
	}

	if templateType == "secondary" {
		flags.EtcdPeers = "http://192.168.0.2"
		flags.EtcdDiscoveryURL = "http://192.168.0.2"
	}

	return flags
}

/* TestCreateIgnitionConfigWithYochu tests that the createIgnitionConfig method
adds a yochu unit to the Ignition config if the yochu version is set. */
func TestCreateIgnitionConfigWithYochu(t *testing.T) {
	for _, templateType := range templateTypes {
		flags := getDefaultTestIgnitionCreateFlags(templateType)

		config, err := createIgnitionConfig(flags)
		if err != nil {
			t.Fatalf("couldn't create %s ignition config with yochu: %s", templateType, err)
		}

		if !strings.Contains(config, "yochu.service") {
			t.Fatalf("yochu service not found in %s ignition config", templateType)
		}
	}
}

/* TestCreateIgnitionConfigWithoutYochu tests that the createIgnitionConfig method
does not add a yochu unit if the yochu version is not set. */
func TestCreateIgnitionConfigWithoutYochu(t *testing.T) {
	for _, templateType := range templateTypes {
		flags := getDefaultTestCreateFlags(templateType)
		flags.YochuVersion = ""

		config, err := createIgnitionConfig(flags)
		if err != nil {
			t.Fatalf("couldn't create %s ignition config with yochu: %s", templateType, err)
		}

		if strings.Contains(config, "yochu.service") {
			t.Fatalf("yochu service found in %s ignition config", templateType)
		}
	}
}
