package swarm

import (
	"fmt"

	"github.com/giantswarm/kocho/provider"
	"github.com/giantswarm/kocho/provider/aws"
)

// ProviderManager describes available and active Providers.
type ProviderManager struct {
	Providers       []ProviderType
	activeProviders []ProviderType
}

// ProviderType describes the type of Provider.
type ProviderType int

// Iota describing the available Providers.
const (
	AWS ProviderType = iota
	OpenStack
	Conair
)

// NewManager returns a new ProviderManager.
func NewManager() ProviderManager {
	return ProviderManager{
		Providers:       []ProviderType{AWS},
		activeProviders: []ProviderType{AWS},
	}
}

// GetByType returns a Provider, given a ProviderType.
func (pm ProviderManager) GetByType(providerType ProviderType) (provider.Provider, error) {
	switch providerType {
	case AWS:
		return aws.Init(), nil
	default:
		return nil, fmt.Errorf("no provider found")
	}
}

// ActiveProviders returns all active Providers.
func (pm ProviderManager) ActiveProviders() ([]provider.Provider, error) {
	var plist []provider.Provider
	for _, providerType := range pm.activeProviders {
		p, err := pm.GetByType(providerType)
		if err != nil {
			return nil, err
		}
		plist = append(plist, p)
	}
	return plist, nil
}
