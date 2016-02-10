package sdk

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

// AutoScalingInstance represents an instance in an AutoScalingGroup.
type AutoScalingInstance struct {
	InstanceId              string
	AvailabilityZone        string
	HealthStatus            string
	LifecycleState          string
	LaunchConfigurationName string
}

// AutoScalingGroup represents an AWS Auto Scaling Group.
type AutoScalingGroup struct {
	Name      string `json:"AutoScalingGroupName"`
	Instances []AutoScalingInstance
}

// NewAutoScaling returns a new AutoScaling.
func NewAutoScaling() *AutoScaling {
	return &AutoScaling{
		client: autoscaling.New(DefaultSessionProvider.GetSession(), AutoScalingConfigs...),
	}
}

// AutoScaling represents the AutoScaling API.
type AutoScaling struct {
	client autoscalingiface.AutoScalingAPI
}

// DescribeAutoScalingGroup returns an AutoScalingGroup, given the name of a group.
func (a AutoScaling) DescribeAutoScalingGroup(name string) (*AutoScalingGroup, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{
			aws.String(name),
		},
		MaxRecords: aws.Int64(1),
	}

	resp, err := a.client.DescribeAutoScalingGroups(input)
	if err != nil {
		return nil, maskAny(err)
	}

	result := AutoScalingGroup{
		Name:      name,
		Instances: make([]AutoScalingInstance, 0, len(resp.AutoScalingGroups[0].Instances)),
	}

	for _, i := range resp.AutoScalingGroups[0].Instances {
		result.Instances = append(result.Instances, AutoScalingInstance{
			InstanceId:              *i.InstanceId,
			AvailabilityZone:        *i.AvailabilityZone,
			HealthStatus:            *i.HealthStatus,
			LifecycleState:          *i.LifecycleState,
			LaunchConfigurationName: *i.LaunchConfigurationName,
		})
	}

	return &result, nil
}

// KillInstance kills the instance in the Auto Scaling Group with the given ID.
func (a AutoScaling) KillInstance(instanceID string) error {
	params := &autoscaling.TerminateInstanceInAutoScalingGroupInput{
		InstanceId:                     aws.String(instanceID),
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}
	_, err := a.client.TerminateInstanceInAutoScalingGroup(params)
	return maskAny(err)
}
