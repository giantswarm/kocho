// Package aws implements a Provider implementation on AWS.
package aws

import (
	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/provider/aws/sdk"
	"github.com/giantswarm/kocho/provider/aws/types"
	"github.com/giantswarm/kocho/swarm/types"

	"github.com/juju/errgo"
)

// AwsProvider represents a Provider running on AWS.
type AwsProvider struct {
	autoscaling    *sdk.AutoScaling
	cloudformation *sdk.CloudFormation
	ec2            *sdk.EC2
	elb            *sdk.ELB
}

const (
	swarmStandaloneTemplate = "standalone"
	swarmSecondaryTemplate  = "secondary"
	swarmPrimaryTemplate    = "primary"
)

// Init initialises the AWS Provider.
func Init() provider.Provider {
	return AwsProvider{
		autoscaling:    sdk.NewAutoScaling(),
		cloudformation: sdk.NewCloudFormation(),
		ec2:            sdk.NewEC2(),
		elb:            sdk.NewELB(),
	}
}

// GetSwarms returns a list of all the Swarms running on AWS.
func (aws AwsProvider) GetSwarms() ([]provider.ProviderSwarm, error) {
	stacks, err := aws.cloudformation.DescribeStacks()
	if err != nil {
		return nil, errgo.Mask(err)
	}

	var swarms []provider.ProviderSwarm
	for _, stack := range stacks.Stacks {
		// for now ignore if there is no stack type yet
		swarmType, _ := findSwarmType(stack.Tags)

		swarms = append(swarms, AwsSwarm{
			Name:         stack.Name,
			Type:         swarmType,
			CreationTime: stack.CreationTime,
			Provider:     aws,
		})
	}

	return swarms, nil
}

// GetSwarm returns a matching Swarm given a name, or ErrNotFound if it cannot be found.
func (aws AwsProvider) GetSwarm(name string) (provider.ProviderSwarm, error) {
	swarm := &AwsSwarm{
		Name:     name,
		Provider: aws,
	}
	stack, err := aws.cloudformation.DescribeStack(name)
	if err != nil {
		if err == provider.ErrNotFound {
			return nil, provider.ErrNotFound
		}
		return nil, errgo.Mask(err)
	}

	swarm.CreationTime = stack.CreationTime

	// for now ignore if there is no stack type yet
	swarm.Type, _ = findSwarmType(stack.Tags)

	return swarm, nil
}

// CreateSwarm creates and returns a Swarm, given a name, CreateFlags and cloud config text.
func (aws AwsProvider) CreateSwarm(name string, flags swarmtypes.CreateFlags, cloudconfigText string) (provider.ProviderSwarm, error) {
	if flags.AWSCreateFlags == nil {
		return nil, errgo.Newf("invalid arguments to create the swarm: AWSCreateFlags must be provided")
	}

	var (
		cloudformationTmpl string
		parametersTmpl     string
		err                error
	)

	switch flags.Type {
	case swarmPrimaryTemplate:
		cloudformationTmpl, err = createPrimaryCloudformationTemplate(name, flags.ClusterSize, flags.TemplateDir, flags.AWSCreateFlags.VPCCIDR)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		parametersTmpl, err = createPrimaryParametersTemplate(flags.ImageURI, cloudconfigText, flags.MachineType, flags.ClusterSize, flags.TemplateDir, flags.AWSCreateFlags)
		if err != nil {
			return nil, errgo.Mask(err)
		}
	case swarmSecondaryTemplate:
		cloudformationTmpl, err = createSecondaryCloudformationTemplate(flags.TemplateDir, flags.AWSCreateFlags.VPCCIDR)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		parametersTmpl, err = createSecondaryParametersTemplate(flags.ImageURI, cloudconfigText, flags.MachineType, flags.CertificateURI, flags.ClusterSize, flags.TemplateDir, flags.AWSCreateFlags)
		if err != nil {
			return nil, errgo.Mask(err)
		}
	case swarmStandaloneTemplate:
		cloudformationTmpl, err = createStandaloneCloudformationTemplate(flags.TemplateDir, flags.AWSCreateFlags.VPCCIDR)
		if err != nil {
			return nil, errgo.Mask(err)
		}
		parametersTmpl, err = createStandaloneParametersTemplate(flags.ImageURI, cloudconfigText, flags.MachineType, flags.CertificateURI, flags.ClusterSize, flags.TemplateDir, flags.AWSCreateFlags)
		if err != nil {
			return nil, errgo.Mask(err)
		}
	}

	_, err = aws.cloudformation.CreateStack(name, flags.Type,
		cloudformationTmpl,
		parametersTmpl,
	)
	if err != nil {
		return nil, errgo.Mask(err)
	}

	return aws.GetSwarm(name)
}

func findSwarmType(tags []types.Tag) (string, error) {
	for _, tag := range tags {
		if tag.Key == "StackType" {
			return tag.Value, nil
		}
	}
	return "", errgo.New("swarm type not found")
}
