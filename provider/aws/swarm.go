package aws

import (
	"fmt"
	"time"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/provider/aws/types"
	"github.com/giantswarm/kocho/swarm/types"
	"github.com/juju/errgo"
)

const (
	statusCreateComplete   = "CREATE_COMPLETE"
	statusRollbackComplete = "ROLLBACK_COMPLETE"
	waitInterval           = 5 * time.Second
)

// AwsSwarm represents a Swarm running on AWS.
type AwsSwarm struct {
	Name         string
	Type         string
	CreationTime time.Time
	Provider     AwsProvider
}

// GetName returns the name of the swarm.
func (s AwsSwarm) GetName() string {
	return s.Name
}

// GetType returns the type of the swarm.
func (s AwsSwarm) GetType() string {
	return s.Type
}

// GetCreationTime returns the time of creation of the swarm.
func (s AwsSwarm) GetCreationTime() time.Time {
	return s.CreationTime
}

// GetStatus returns the status, and a status reason, of the swarm.
func (s AwsSwarm) GetStatus() (string, string, error) {
	stack, err := s.Provider.cloudformation.DescribeStack(s.Name)
	if err != nil {
		return "", "", err
	}

	return stack.Status, stack.StatusReason, nil
}

// GetPublicDNS returns the public DNS address of the swarm.
func (s AwsSwarm) GetPublicDNS() (string, error) {
	lbResource, err := s.getPublicLoadBalancer()
	if err != nil {
		return "", err
	}

	lb, err := s.Provider.elb.DescribeLoadBalancer(lbResource.PhysicalId)
	if err != nil {
		return "", err
	}

	return lb.DNSName, nil
}

// GetPrivateDNS returns the private DNS address of the swarm.
func (s AwsSwarm) GetPrivateDNS() (string, error) {
	lbResource, err := s.getPrivateLoadBalancer()
	if err != nil {
		return "", err
	}

	lb, err := s.Provider.elb.DescribeLoadBalancer(lbResource.PhysicalId)
	if err != nil {
		return "", err
	}

	return lb.DNSName, nil
}

// GetInstances returns all the instances of the swarm.
func (s AwsSwarm) GetInstances() ([]swarmtypes.Instance, error) {
	awsInstances, err := s.Provider.ec2.FindInstancesByTags(types.Tag{
		Key:   cloudFormationStackTag,
		Value: s.Name,
	})
	if err != nil {
		return nil, err
	}

	var instances []swarmtypes.Instance
	for _, awsInstance := range awsInstances {
		instances = append(instances, swarmtypes.Instance{
			Id:               awsInstance.InstanceId,
			Image:            awsInstance.ImageId,
			Type:             awsInstance.InstanceType,
			PublicIPAddress:  awsInstance.PublicIPAddress,
			PrivateIPAddress: awsInstance.PrivateIPAddress,
			PublicDNSName:    awsInstance.PublicDNSName,
			PrivateDNSName:   awsInstance.PrivateDNSName,
		})
	}

	return instances, nil
}

// KillInstance kills the given instance in the swarm.
func (s AwsSwarm) KillInstance(i swarmtypes.Instance) error {
	if err := s.Provider.killInstance(i.Id); err != nil {
		return errgo.Mask(err)
	}
	return nil
}

// Destroy destroys the swarm.
func (s AwsSwarm) Destroy() error {
	return s.Provider.cloudformation.DeleteStack(s.Name)
}

// WaitUntil waits until the swarm is in the given state.
func (s AwsSwarm) WaitUntil(status string) error {
	switch status {
	case provider.StatusCreated:
		err := s.waitForCompletion()
		if err != nil {
			return err
		}

		if s.Type == "primary" {
			return s.waitForPrivateLoadBalancer()
		} else {
			return s.waitForAutoScaler()
		}
	case provider.StatusDeleted:
		return s.waitForDeletion()
	default:
		return fmt.Errorf("waiting for status '%s' is not implemented yet.", status)
	}
}

func (s AwsSwarm) waitForCompletion() error {
	for {
		status, _, err := s.GetStatus()
		if err != nil {
			return err
		}

		if status == statusCreateComplete {
			return nil // success
		}

		if status == statusRollbackComplete {
			return fmt.Errorf("swarm was rolled back. Please check AWS Console for error details")
		}

		time.Sleep(waitInterval)
	}
	return nil
}

func (s AwsSwarm) waitForDeletion() error {
	_, _, err := s.GetStatus()
	if err == provider.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	time.Sleep(waitInterval)
	return s.waitForDeletion()
}

func (s AwsSwarm) waitForAutoScaler() error {
	as, err := s.getAutoScaler()
	if err != nil {
		return err
	}

	if as.Status != statusCreateComplete {
		fmt.Println(as.Status, err)
		time.Sleep(waitInterval)
		if err := s.waitForAutoScaler(); err != nil {
			return err
		}
	}
	return nil
}

func (s AwsSwarm) waitForPrivateLoadBalancer() error {
	elb, err := s.getPrivateLoadBalancer()
	if err != nil {
		return err
	}

	if elb.Status != statusCreateComplete {
		fmt.Println(elb.Status, err)
		time.Sleep(waitInterval)
		if err := s.waitForPrivateLoadBalancer(); err != nil {
			return err
		}
	}
	return nil
}

func (s AwsSwarm) getResources() ([]types.StackResource, error) {
	resources, err := s.Provider.cloudformation.DescribeStackResources(s.Name)
	if err != nil {
		return nil, err
	}

	return resources.StackResources, nil
}

func (s AwsSwarm) findResourcesByType(resourceType string) ([]types.StackResource, error) {
	resources, err := s.getResources()
	if err != nil {
		return nil, err
	}

	var foundResources []types.StackResource
	for _, resource := range resources {
		if resource.Type == resourceType {
			foundResources = append(foundResources, resource)
		}
	}
	return foundResources, nil
}

func (s AwsSwarm) findResourcesByLogicalId(logicalId string) ([]types.StackResource, error) {
	resources, err := s.getResources()
	if err != nil {
		return nil, err
	}

	var foundResources []types.StackResource
	for _, resource := range resources {
		if resource.LogicalId == logicalId {
			foundResources = append(foundResources, resource)
		}
	}
	return foundResources, nil
}

func (s AwsSwarm) getAutoScaler() (*types.StackResource, error) {
	resources, err := s.findResourcesByType("AWS::AutoScaling::AutoScalingGroup")
	if err != nil {
		return nil, err
	}

	if len(resources) < 1 {
		return nil, fmt.Errorf("autoscaler not found")
	}
	return &resources[0], nil
}

func (s AwsSwarm) getPublicLoadBalancer() (*types.StackResource, error) {
	resources, err := s.findResourcesByLogicalId("ElasticLoadBalancerPublic")
	if err != nil {
		return nil, err
	}

	if len(resources) < 1 {
		return nil, fmt.Errorf("public LoadBalancer not found")
	}
	return &resources[0], nil
}

func (s AwsSwarm) getPrivateLoadBalancer() (*types.StackResource, error) {
	resources, err := s.findResourcesByLogicalId("ElasticLoadBalancerPrivate")
	if err != nil {
		return nil, err
	}

	if len(resources) < 1 {
		return nil, fmt.Errorf("private LoadBalancer not found")
	}
	return &resources[0], nil
}
