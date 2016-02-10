package swarm

import (
	"strings"
	"testing"

	"github.com/giantswarm/kocho/swarm/types"
)

var (
	templateTypes = []string{"primary", "secondary", "standalone"}
)

// getDefaultTestCreateFlags returns a CreateFlags that can be used for testing
func getDefaultTestCreateFlags(templateType string) swarmtypes.CreateFlags {
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
		EtcdDiscoveryURL: "",
		TemplateDir:      "../default-templates", // Hack to use default built in templates
		MachineType:      "m3.large",
		ImageURI:         "ami-5k2l4639",
		CertificateURI:   "arn:aws:iam::063170442734:server-certificate/wildcard-2015-chained",
		AWSCreateFlags:   &awsCreateFlags,
	}

	if templateType == "secondary" {
		flags.EtcdPeers = "http://192.168.0.2"
		flags.EtcdDiscoveryURL = "http://192.168.0.2"
	}

	return flags
}

/* TestCreateCloudConfigWithYochu tests that the createCloudConfig method
adds a yochu unit to the cloud config if the yochu version is set. */
func TestCreateCloudConfigWithYochu(t *testing.T) {
	for _, templateType := range templateTypes {
		flags := getDefaultTestCreateFlags(templateType)

		config, err := createCloudConfig(flags)
		if err != nil {
			t.Fatalf("couldn't create %s cloud config with yochu: %s", templateType, err)
		}

		if !strings.Contains(config, "yochu.service") {
			t.Fatalf("yochu service not found in %s cloud config", templateType)
		}
	}
}

/* TestCreateCloudConfigWithoutYochu tests that the createCloudConfig method
does not add a yochu unit if the yochu version is not set. */
func TestCreateCloudConfigWithoutYochu(t *testing.T) {
	for _, templateType := range templateTypes {
		flags := getDefaultTestCreateFlags(templateType)
		flags.YochuVersion = ""

		config, err := createCloudConfig(flags)
		if err != nil {
			t.Fatalf("couldn't create %s cloud config with yochu: %s", templateType, err)
		}

		if strings.Contains(config, "yochu.service") {
			t.Fatalf("yochu service found in %s cloud config", templateType)
		}
	}
}
