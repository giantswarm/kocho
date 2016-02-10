// Package swarm provides general business logic of handling swarms.
package swarm

import (
	"time"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/swarm/types"
)

// Swarm represents a cluster of CoreOS machines.
type Swarm struct {
	Name     string
	Type     string
	Created  time.Time
	provider provider.ProviderSwarm
}

func createSwarm(swarm provider.ProviderSwarm) *Swarm {
	return &Swarm{
		Name:     swarm.GetName(),
		Type:     swarm.GetType(),
		Created:  swarm.GetCreationTime(),
		provider: swarm,
	}
}

// WaitUntil waits till the Swarm reaches a given status.
func (s *Swarm) WaitUntil(status string) error {
	return s.provider.WaitUntil(status)
}

// GetStatus returns the status, and a possibly empty status reason, of the Swarm.
func (s *Swarm) GetStatus() (string, string, error) {
	return s.provider.GetStatus()
}

// GetInstances returns all the instances of the Swarm.
func (s *Swarm) GetInstances() ([]swarmtypes.Instance, error) {
	return s.provider.GetInstances()
}

// GetPublicDNS returns the public DNS address of the Swarm.
func (s *Swarm) GetPublicDNS() (string, error) {
	return s.provider.GetPublicDNS()
}

// GetPrivateDNS returns the private DNS address of the Swarm.
func (s *Swarm) GetPrivateDNS() (string, error) {
	return s.provider.GetPrivateDNS()
}

// KillInstance kills a given instance of the Swarm.
func (s *Swarm) KillInstance(i swarmtypes.Instance) error {
	return s.provider.KillInstance(i)
}

// Destroy destroys the Swarm.
func (s *Swarm) Destroy() error {
	return s.provider.Destroy()
}
