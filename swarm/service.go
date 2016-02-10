package swarm

import (
	"github.com/juju/errgo"

	"github.com/giantswarm/kocho/swarm/types"
)

// Config describes the configuration of a Service.
type Config struct {
}

// Dependencies describe the dependencies of a Service.
type Dependencies struct {
}

// NewService returns a new Service.
func NewService(cfg Config, deps Dependencies) *Service {
	return &Service{
		Config:       cfg,
		Dependencies: deps,

		providers: NewManager(),
	}
}

// Service descibes a ProviderManager, and some configuration and dependencies.
type Service struct {
	Config
	Dependencies

	providers ProviderManager
}

// Create creates and returns a Swarm, given a name for the swarm, a ProviderType, and CreateFlags.
func (srv *Service) Create(name string, providerType ProviderType, flags swarmtypes.CreateFlags) (*Swarm, error) {
	p, err := srv.providers.GetByType(providerType)
	if err != nil {
		return nil, err
	}

	switch providerType {
	case AWS:
		if flags.AWSCreateFlags == nil {
			return nil, errgo.Newf("AWSCreateFlags must be provided")
		}
	}

	cfg, err := createCloudConfig(flags)
	if err != nil {
		return nil, err
	}
	swarm, err := p.CreateSwarm(name, flags, cfg)
	if err != nil {
		return nil, err
	}

	return createSwarm(swarm), nil
}

// List returns all available Swarms.
func (srv *Service) List() ([]*Swarm, error) {
	plist, err := srv.providers.ActiveProviders()
	if err != nil {
		return nil, err
	}

	var swarms []*Swarm
	for _, provider := range plist {
		swarmList, err := provider.GetSwarms()
		if err != nil {
			return nil, err
		}

		for _, swarm := range swarmList {
			swarms = append(swarms, createSwarm(swarm))
		}
	}

	return swarms, nil
}

// Get returns a Swarm, given a swarm name, and a ProviderType.
func (srv *Service) Get(name string, providerType ProviderType) (*Swarm, error) {
	p, err := srv.providers.GetByType(providerType)
	if err != nil {
		return nil, err
	}

	swarm, err := p.GetSwarm(name)
	if err != nil {
		return nil, err
	}

	return createSwarm(swarm), nil
}
