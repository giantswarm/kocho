package dns

import (
	"github.com/giantswarm/kocho/swarm"
)

// NoopDNS provides a DNSService implementation that does nothing.
// Useful for when we don't want to set up DNS at all.
type NoopDNS struct{}

// NewNoopDNS returns a new NoopDNS.
func NewNoopDNS() *NoopDNS {
	return &NoopDNS{}
}

// createSwarmEntries does nothing, returning immediately.
func (ndns *NoopDNS) createSwarmEntries(s *swarm.Swarm, entries *Entries) error {
	return nil
}

// deleteEntries does nothing, returning immediately.
func (ndns *NoopDNS) deleteEntries(name string, entries *Entries) error {
	return nil
}

// update does nothing, returning immediately.
func (ndns *NoopDNS) update(stackName, cname, dns string, entries *Entries) error {
	return nil
}
