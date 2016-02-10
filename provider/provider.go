// Package provider defines the cloud backend for the `swarm` package.
package provider

import (
	"errors"
	"time"

	"github.com/giantswarm/kocho/swarm/types"
)

var (
	ErrNotFound = errors.New("not found")
)

const (
	StatusCreated = "created"
	StatusDeleted = "deleted"
)

// ProviderSwarm represents a Swarm running in a Provider.
type ProviderSwarm interface {
	GetName() string
	GetType() string
	GetCreationTime() time.Time
	GetStatus() (string, string, error)
	GetPublicDNS() (string, error)
	GetPrivateDNS() (string, error)
	GetInstances() ([]swarmtypes.Instance, error)
	WaitUntil(string) error
	KillInstance(swarmtypes.Instance) error
	Destroy() error
}

// Provider represents a system that can manage Swarm.
type Provider interface {
	CreateSwarm(name string, flags swarmtypes.CreateFlags, cloudconfigText string) (ProviderSwarm, error)
	GetSwarm(name string) (ProviderSwarm, error)
	GetSwarms() ([]ProviderSwarm, error)
}
