package sdk

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/provider/aws/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
)

// Stacks represents a list containing multiple Stack.
type Stacks struct {
	Stacks []Stack
}

// Stack represents a CloudFormation stack.
type Stack struct {
	Id           string `json:"StackId"`
	Name         string `json:"StackName"`
	Status       string `json:"StackStatus"`
	StatusReason string `json:"StackStatusReason"`
	Tags         []types.Tag
	CreationTime time.Time
}

// StackResources represents a list containing multiple StackResource.
type StackResources struct {
	StackResources []types.StackResource
}

// NewCloudFormation returns a new CloudFormation.
func NewCloudFormation() *CloudFormation {
	return &CloudFormation{
		client: cloudformation.New(DefaultSessionProvider.GetSession(), CloudFormationConfigs...),
	}
}

// CloudFormation represents the CloudFormation API.
type CloudFormation struct {
	client cloudformationiface.CloudFormationAPI
}

// CreateStack creates a CloudFormation stack, given a name, a type of stack, and template and parameters files.
func (c CloudFormation) CreateStack(name, stackType, templateFile, parametersFile string) (*Stack, error) {
	var awsParameters []*cloudformation.Parameter
	if err := c.loadFile(parametersFile, &awsParameters); err != nil {
		return nil, err
	}

	var templateBody string
	if data, err := ioutil.ReadFile(templateFile); err != nil {
		return nil, err
	} else {
		templateBody = string(data)
	}

	input := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		TemplateBody: aws.String(templateBody),
		Parameters:   awsParameters,
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("StackType"),
				Value: aws.String(stackType),
			},
		},
	}

	resp, err := c.client.CreateStack(input)
	if err != nil {
		return nil, err
	}

	stacks, err := c.describeStacks(&cloudformation.DescribeStacksInput{
		StackName: resp.StackId,
	})
	return &stacks.Stacks[0], err
}

// DescribeStack returns a Stack, given the name of a CloudFormation stack.
func (c CloudFormation) DescribeStack(name string) (*Stack, error) {
	input := cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	stacks, err := c.describeStacks(&input)
	if err != nil {
		return nil, err
	}
	return &stacks.Stacks[0], nil
}

// DescribeStacks returns all Stacks.
func (c CloudFormation) DescribeStacks() (*Stacks, error) {
	stacks, err := c.describeStacks(nil)
	return stacks, err
}

// DescribeStackResources returns the StackResources, given the name of a stack.
func (c CloudFormation) DescribeStackResources(name string) (*StackResources, error) {
	resp, err := c.client.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	resources := &StackResources{
		StackResources: make([]types.StackResource, 0, len(resp.StackResources)),
	}
	for _, detail := range resp.StackResources {
		resources.StackResources = append(resources.StackResources,
			types.StackResource{
				Type:       *detail.ResourceType,
				Status:     *detail.ResourceStatus,
				PhysicalId: *detail.PhysicalResourceId,
				LogicalId:  *detail.LogicalResourceId,
			},
		)
	}
	return resources, nil
}

// DeleteStack deletes the CloudFormation stack of the given name.
func (c CloudFormation) DeleteStack(name string) error {
	_, err := c.client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	})
	if err != nil {
		return err
	}
	return nil
}

func (c CloudFormation) describeStacks(input *cloudformation.DescribeStacksInput) (*Stacks, error) {
	resp, err := c.client.DescribeStacks(input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ValidationError" {
			return nil, provider.ErrNotFound
		}
		return nil, err
	}

	result := &Stacks{Stacks: []Stack{}}
	for _, awsStack := range resp.Stacks {
		stack := Stack{
			Id:           *awsStack.StackId,
			Name:         *awsStack.StackName,
			Status:       *awsStack.StackStatus,
			Tags:         fromCloudFormationTags(awsStack.Tags),
			CreationTime: *awsStack.CreationTime,
		}

		if awsStack.StackStatusReason != nil {
			stack.StatusReason = *awsStack.StackStatusReason
		}

		result.Stacks = append(result.Stacks, stack)
	}
	return result, nil
}

func (c CloudFormation) loadFile(file string, target interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func fromCloudFormationTags(tags []*cloudformation.Tag) []types.Tag {
	result := make([]types.Tag, 0, len(tags))

	for _, tag := range tags {
		result = append(result, types.Tag{
			Key:   *tag.Key,
			Value: *tag.Value,
		})
	}

	return result
}
