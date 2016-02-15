package sdk

import (
	"github.com/giantswarm/kocho/provider/aws/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

const (
	ec2StateRunning = "running"
)

// NewEC2 returns a new EC2.
func NewEC2() *EC2 {
	return &EC2{
		client: ec2.New(DefaultSessionProvider.GetSession(), EC2Configs...),
	}
}

// EC2 represents the EC2 API.
type EC2 struct {
	client ec2iface.EC2API
}

// DescribeInstances returns a list of Instances, given a list of instance IDs.
func (e EC2) DescribeInstances(instanceIds []string) ([]types.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{},
		MaxResults:  aws.Int64(int64(len(instanceIds))),
	}
	for _, id := range instanceIds {
		params.InstanceIds = append(params.InstanceIds, aws.String(id))
	}
	return e.describeInstances(params)
}

// FindInstancesByTags returns a list of Instances, given a list of Tags.
func (e EC2) FindInstancesByTags(tags ...types.Tag) ([]types.Instance, error) {
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{},
	}
	for _, tag := range tags {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name:   aws.String("tag:" + tag.Key),
			Values: []*string{aws.String(tag.Value)},
		})
	}
	return e.describeInstances(params)
}

func (e EC2) describeInstances(input *ec2.DescribeInstancesInput) ([]types.Instance, error) {
	resp, err := e.client.DescribeInstances(input)
	if err != nil {
		return nil, maskAny(err)
	}

	result := make([]types.Instance, 0)
	for _, reservation := range resp.Reservations {
		for _, i := range reservation.Instances {
			// For now we filter everything that is now running, as certain
			// consumers of this expect all returned instances to be good instances
			if *(i.State.Name) != ec2StateRunning {
				continue
			}
			inst := types.Instance{
				InstanceId:       *i.InstanceId,
				ImageId:          *i.ImageId,
				InstanceType:     *i.InstanceType,
				PublicIPAddress:  *i.PublicIpAddress,
				PublicDNSName:    *i.PublicDnsName,
				PrivateIPAddress: *i.PrivateIpAddress,
				PrivateDNSName:   *i.PrivateDnsName,
			}

			result = append(result, inst)
		}
	}
	return result, nil
}
