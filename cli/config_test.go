package cli

import (
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/giantswarm/kocho/swarm/types"
)

func assertEquals(expected, actual string) bool {
	return expected == actual
}

func clusterType(obj *swarmtypes.CreateFlags) string {
	return obj.Type
}

func machineType(obj *swarmtypes.CreateFlags) string {
	return obj.MachineType
}

func awsKeypairName(obj *swarmtypes.CreateFlags) string {
	return obj.AWSCreateFlags.KeypairName
}

var testData = []struct {
	Getter      func(*swarmtypes.CreateFlags) string
	Config      string
	PflagArgs   []string
	Expected    string
	Description string
}{
	//  Getter	  Config			Value		Pflag			Value		Expected	Desc
	{clusterType, "", nil, "standalone", "type: pflag type has a default value"},
	{clusterType, "type: primary", nil, "primary", "type: Config files can overwrite default"},
	{clusterType, "", []string{"--type=secondary"}, "secondary", "type: CLI args also overwrite defaults"},
	{clusterType, "type: primary", []string{"--type=secondary"}, "secondary", "type: CLI flags win over config file"},

	{machineType, "", nil, "m3.large", "machine-type: pflag machine-type has a default value"},
	{machineType, "machine-type: x3.xlarge", nil, "x3.xlarge", "machine-type: config file is used, when no flag is given"},
	{machineType, "machine-type: x3.xlarge", []string{"--machine-type=t2.micro"}, "t2.micro", "machine-type: CLI wins over config file"},

	{awsKeypairName, "aws-keypair: aws-config", nil, "aws-config", "keypair is read from config"},
	{awsKeypairName, "aws-keypair: aws-config", []string{"--aws-keypair=cli"}, "cli", "Keypair is configurable via CLI"},
}

func TestNewViperCreateFlags(t *testing.T) {
	for index, record := range testData {
		config := NewConfig()
		config.SetConfigType("yaml")

		if record.Config != "" {
			if err := config.ReadConfig(strings.NewReader(record.Config + "\n")); err != nil {
				t.Errorf("test %d: Invalid config: %v", index, err)
				continue
			}
		}

		f := pflag.NewFlagSet("test", pflag.ExitOnError)
		registerCreateFlags(f)
		config.configFromPFlags(f)

		if record.PflagArgs != nil {
			if err := f.Parse(record.PflagArgs); err != nil {
				t.Errorf("test %d: Failed to parse args: %v", err)
				continue
			}
		}

		obj := config.newViperCreateFlags()
		actual := record.Getter(&obj)
		if !assertEquals(record.Expected, actual) {
			t.Errorf("test %d: %s\n\tExpected '%s', got '%s'", index, record.Description, record.Expected, actual)
		}
	}
}
