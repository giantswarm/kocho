package sdk

import (
	"github.com/aws/aws-sdk-go/aws"
	//	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elb/elbiface"
)

// LoadBalancer represents a load balancer running in AWS.
type LoadBalancer struct {
	LoadBalancerName string
	DNSName          string
	Scheme           string
}

// NewELB returns a new ELB.
func NewELB() *ELB {
	return &ELB{
		client: elb.New(DefaultSessionProvider.GetSession(), ELBConfigs...),
	}
}

// ELB represents the ELB API.
type ELB struct {
	client elbiface.ELBAPI
}

// DescribeLoadBalancer returns a LoadBalancer, given a matching name.
func (e ELB) DescribeLoadBalancer(name string) (*LoadBalancer, error) {
	resp, err := e.client.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{
		LoadBalancerNames: []*string{aws.String(name)},
	})
	if err != nil {
		return nil, maskAny(err)
	}
	lb := LoadBalancer{
		LoadBalancerName: *resp.LoadBalancerDescriptions[0].LoadBalancerName,
		DNSName:          *resp.LoadBalancerDescriptions[0].DNSName,
		Scheme:           *resp.LoadBalancerDescriptions[0].Scheme,
	}
	return &lb, nil
}
