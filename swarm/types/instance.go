package swarmtypes

import (
	"github.com/juju/errgo"
)

// Instance describes a machine running in a swarm.
type Instance struct {
	Id               string
	Image            string
	Type             string
	PublicIPAddress  string
	PublicDNSName    string
	PrivateIPAddress string
	PrivateDNSName   string
}

// FilterInstanceById filters an instance from an existing slice by its id.
func FilterInstanceById(instances []Instance, instanceID string) []Instance {
	filteredInstances := make([]Instance, 0)
	for _, i := range instances {
		if i.Id != instanceID {
			filteredInstances = append(filteredInstances, i)
		}
	}
	return filteredInstances
}

// FindInstanceById searches for an instance in a slice by its id.
func FindInstanceById(instances []Instance, instanceID string) (Instance, error) {
	for _, i := range instances {
		if i.Id == instanceID {
			return i, nil
		}
	}
	return Instance{}, errgo.Newf("couldn't find instance %s", instanceID)
}
