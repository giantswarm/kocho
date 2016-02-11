// Package dns provides the capability to modify DNS records for swarms.
package dns

import (
	"bytes"
	"text/template"

	"github.com/giantswarm/kocho/swarm"
	"github.com/giantswarm/kocho/swarm/types"
)

// NamingPattern defines the template to use when generating the domain names to create.
// Assume the value 'Stack' can be used to reference the cluster name.
//
// The generated domain here will always be postfixed with the zone, so this is not part of the template.
type NamingPattern struct {
	// The name of the zone to create the domains in
	Zone string

	// Golang templates
	Catchall        string
	CatchallPrivate string
	Public          string
	Private         string
	Fleet           string
}

var (
	DefaultNamingPattern = NamingPattern{
		Zone:            "example.com",
		Catchall:        "*.{{.Stack}}",
		CatchallPrivate: "*.{{.Stack}}.private",
		Public:          "{{.Stack}}",
		Private:         "{{.Stack}}.private",
		Fleet:           "{{.Stack}}.fleet",
	}
)

// GetEntries returns Entries, given a stack name.
func (np NamingPattern) GetEntries(stackName string) *Entries {
	return &Entries{
		Zone: np.Zone,

		Catchall:        np.mustParse(np.Catchall, stackName),
		CatchallPrivate: np.mustParse(np.CatchallPrivate, stackName),
		Public:          np.mustParse(np.Public, stackName),
		Private:         np.mustParse(np.Private, stackName),
		Fleet:           np.mustParse(np.Fleet, stackName),
	}
}

// Entries represents a set of DNS entries.
type Entries struct {
	// The root domain for all entries, e.g. giantswarm.io or giantswarm.co.uk
	Zone string

	// Endpoint domains
	Catchall        string
	CatchallPrivate string
	Public          string
	Private         string
	Fleet           string
}

// DNSService provides mangement of a swarm's DNS entries.
type DNSService interface {
	createSwarmEntries(s *swarm.Swarm, entries *Entries) error
	deleteEntries(name string, entries *Entries) error
	update(stackName, cname, dns string, entries *Entries) error
}

// CreateSwarmEntries creates DNS entries, given a NamingPattern and Swarm.
func CreateSwarmEntries(service DNSService, pattern NamingPattern, s *swarm.Swarm) error {
	return service.createSwarmEntries(s, pattern.GetEntries(s.Name))
}

// DeleteEntries deletes DNS entries, given a NamingPattern and stack name.
func DeleteEntries(service DNSService, pattern NamingPattern, stackName string) error {
	return service.deleteEntries(stackName, pattern.GetEntries(stackName))
}

// Update ensures that necessary DNS records that might have changed are updated.
//
// NOTE: This currently only checks the fleet entry.
// Returns false if no update was necessary.
func Update(service DNSService, pattern NamingPattern, s *swarm.Swarm, instances []swarmtypes.Instance) (bool, error) {
	dnsEntries := pattern.GetEntries(s.Name)
	dnsChanged := false

	for _, inst := range instances {
		if inst.PublicDNSName == "" {
			continue
		}

		if err := service.update(s.Name, dnsEntries.Fleet, inst.PublicDNSName, dnsEntries); err != nil {
			return true, err
		}
		dnsChanged = true
	}

	return dnsChanged, nil
}

func (np NamingPattern) parse(templateText, stackName string) (string, error) {
	t, err := template.New("naming-pattern").Parse(templateText)
	if err != nil {
		return "", err
	}

	buffer := bytes.NewBufferString("")
	if err := t.Execute(buffer, map[string]string{
		"Stack": stackName,
	}); err != nil {
		return "", err
	}
	return buffer.String() + "." + np.Zone, nil
}

func (np NamingPattern) mustParse(template, stackName string) string {
	s, err := np.parse(template, stackName)
	if err != nil {
		panic(err.Error())
	}
	return s
}
