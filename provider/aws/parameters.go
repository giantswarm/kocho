package aws

import (
	"encoding/base64"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/juju/errgo"

	"github.com/giantswarm/kocho/swarm/types"
)

const (
	generatedParametersPath = "/tmp/aws-parameters.json"

	primaryParametersTemplateName    = "primary-parameters.tmpl"
	secondaryParametersTemplateName  = "secondary-parameters.tmpl"
	standaloneParametersTemplateName = "standalone-parameters.tmpl"
)

type parameters struct {
	CloudConfig    string
	SSLCertificate string
	InstanceType   string
	ClusterSize    int
	KeyPair        string
	VpcId          string
	Subnet         string
	AZ             string
	AmiId          string
}

func createPrimaryParametersTemplate(image, cloudConfig, machineType string, clusterSize int, templateDir string, awsFlags *swarmtypes.AWSCreateFlags) (string, error) {
	parametersTemplatePath := path.Join(templateDir, primaryParametersTemplateName)

	p := parameters{
		CloudConfig:  base64.StdEncoding.EncodeToString([]byte(cloudConfig)),
		InstanceType: machineType,
		ClusterSize:  clusterSize,
		KeyPair:      awsFlags.KeypairName,
		VpcId:        awsFlags.VPC,
		Subnet:       awsFlags.Subnet,
		AZ:           awsFlags.AvailabilityZone,
		AmiId:        image,
	}

	return parseParametersTemplate(parametersTemplatePath, p)
}

func createSecondaryParametersTemplate(image, cloudConfig, machineType, certificate string, clusterSize int, templateDir string, awsFlags *swarmtypes.AWSCreateFlags) (string, error) {
	parametersTemplatePath := path.Join(templateDir, secondaryParametersTemplateName)

	p := parameters{
		CloudConfig:    base64.StdEncoding.EncodeToString([]byte(cloudConfig)),
		SSLCertificate: certificate,
		InstanceType:   machineType,
		ClusterSize:    clusterSize,
		KeyPair:        awsFlags.KeypairName,
		VpcId:          awsFlags.VPC,
		Subnet:         awsFlags.Subnet,
		AZ:             awsFlags.AvailabilityZone,
		AmiId:          image,
	}

	return parseParametersTemplate(parametersTemplatePath, p)
}

func createStandaloneParametersTemplate(image, cloudConfig, machineType, certificate string, clusterSize int, templateDir string, awsFlags *swarmtypes.AWSCreateFlags) (string, error) {
	parametersTemplatePath := path.Join(templateDir, standaloneParametersTemplateName)

	p := parameters{
		CloudConfig:    base64.StdEncoding.EncodeToString([]byte(cloudConfig)),
		SSLCertificate: certificate,
		InstanceType:   machineType,
		ClusterSize:    clusterSize,
		KeyPair:        awsFlags.KeypairName,
		VpcId:          awsFlags.VPC,
		Subnet:         awsFlags.Subnet,
		AZ:             awsFlags.AvailabilityZone,
		AmiId:          image,
	}

	return parseParametersTemplate(parametersTemplatePath, p)
}

func parseParametersTemplate(templatePath string, p parameters) (string, error) {
	f, err := os.Create(generatedParametersPath)
	if err != nil {
		return generatedParametersPath, errgo.Mask(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	absoluteParametersTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	templateData, err := ioutil.ReadFile(absoluteParametersTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	var tmpl *template.Template
	if tmpl, err = template.New("cfg").Parse(string(templateData)); err != nil {
		return generatedParametersPath, errgo.Mask(err)
	}

	if err = tmpl.Execute(f, p); err != nil {
		return generatedParametersPath, errgo.Mask(err)
	}

	return generatedParametersPath, nil
}
