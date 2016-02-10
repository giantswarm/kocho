package aws

import (
	"github.com/giantswarm/kocho/provider/aws/types"
)

func (aws AwsProvider) getInstances(autoScalingGroupName string) ([]types.Instance, error) {
	autoScalingGroup, err := aws.autoscaling.DescribeAutoScalingGroup(autoScalingGroupName)

	var instanceIds []string
	for _, i := range autoScalingGroup.Instances {
		instanceIds = append(instanceIds, i.InstanceId)
	}
	instances, err := aws.ec2.DescribeInstances(instanceIds)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

func (aws AwsProvider) killInstance(instanceID string) error {
	return aws.autoscaling.KillInstance(instanceID)
}
