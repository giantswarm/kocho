package cli

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/giantswarm/kocho/dns"
	"github.com/giantswarm/kocho/swarm/types"
)

var (
	// ConfigHomePath defines where to load configuration files from.
	ConfigHomePath = os.Getenv("HOME") + "/.giantswarm/"
)

// NewConfig returns a new KochoConfiguration.
func NewConfig() *KochoConfiguration {
	return &KochoConfiguration{
		viper.New(),
	}
}

// KochoConfiguration holds Viper configuration.
type KochoConfiguration struct {
	*viper.Viper
}

func (viper *KochoConfiguration) configFromPFlags(set *pflag.FlagSet) error {
	// Copy command specific flags into viper
	if err := viper.BindPFlags(set); err != nil {
		return err
	}
	return nil
}

func (viper *KochoConfiguration) loadConfig() error {
	viper.SetConfigName("kocho")
	// Prefer loading from the current working directory, but fallback to $HOME/.giantswarm
	viper.AddConfigPath(".")
	viper.AddConfigPath(ConfigHomePath)
	return viper.ReadInConfig()
}

// getViperCreateFlags creates a swarmtypes.CreateFlags from the viper config default values.
func (viper *KochoConfiguration) newViperCreateFlags() swarmtypes.CreateFlags {
	return swarmtypes.CreateFlags{
		Type:        viper.GetString("type"),
		TemplateDir: viper.GetString("template-dir"),

		Tags:             viper.GetString("tags"),
		YochuVersion:     viper.GetString("yochu"),
		FleetVersion:     viper.GetString("fleet-version"),
		EtcdPeers:        viper.GetString("etcd-peers"),
		EtcdVersion:      viper.GetString("etcd-version"),
		EtcdDiscoveryURL: viper.GetString("etcd-discovery-url"),
		DockerVersion:    viper.GetString("docker-version"),
		ClusterSize:      viper.GetInt("cluster-size"),

		// Provider interpreted
		ImageURI:       viper.GetString("image"),
		MachineType:    viper.GetString("machine-type"),
		CertificateURI: viper.GetString("certificate"),

		AWSCreateFlags: &swarmtypes.AWSCreateFlags{
			KeypairName:      viper.GetString("aws-keypair"),
			Subnet:           viper.GetString("aws-subnet"),
			VPC:              viper.GetString("aws-vpc"),
			VPCCIDR:          viper.GetString("aws-vpc-cidr"),
			AvailabilityZone: viper.GetString("aws-az"),
		},
	}
}

func (viper *KochoConfiguration) getDNSNamingPattern() dns.NamingPattern {
	return dns.NamingPattern{
		Zone:            viper.GetString("dns-zone"),
		Catchall:        viper.GetString("dns-catchall"),
		CatchallPrivate: viper.GetString("dns-catchall-private"),
		Public:          viper.GetString("dns-public"),
		Private:         viper.GetString("dns-private"),
		Fleet:           viper.GetString("dns-fleet"),
	}
}
