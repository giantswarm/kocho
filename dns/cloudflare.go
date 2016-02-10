package dns

import (
	"fmt"
	"os"
	"sync"

	"github.com/giantswarm/kocho/swarm"

	"github.com/crackcomm/cloudflare"
	"github.com/juju/errgo"
	"golang.org/x/net/context"
)

// CloudFlareDNS represents a client to the CloudFlare API.
type CloudFlareDNS struct {
	mutex   sync.Mutex
	_client *cloudflare.Client
}

// CreateSwarmEntries creates DNS entries, given a Swarm and Entries to create.
func (cli *CloudFlareDNS) CreateSwarmEntries(s *swarm.Swarm, e *Entries) error {
	ctx := context.TODO()

	zone, err := cli.findZone(ctx, e.Zone)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	privateDns, err := s.GetPrivateDNS()
	if err != nil {
		return fmt.Errorf("Couldn't get private dns for swarm: %s - %v", s.Name, err)
	}
	instances, err := s.GetInstances()
	if err != nil {
		return fmt.Errorf("Failed fetch list of instances for swarm: %s - %v", s.Name, err)
	}
	if len(instances) < 1 {
		return fmt.Errorf("Couldn't get swarm instances of: %s - %v", s.Name, err)
	}

	if s.Type != "primary" {
		publicDns, err := s.GetPublicDNS()
		if err != nil {
			return fmt.Errorf("Couldn't get public dns for swarm: %s - %v", s.Name, err)
		}
		if err := cli.createRecord(ctx, zone.ID, e.Catchall, publicDns); err != nil {
			return fmt.Errorf("Couldn't create catchall dns entry: %s %s - %v", e.Catchall, publicDns, err)
		}
		if err := cli.createRecord(ctx, zone.ID, e.Public, publicDns); err != nil {
			return fmt.Errorf("Couldn't create public dns entry: %s %s - %v", e.Public, publicDns, err)
		}
	}

	if err := cli.createRecord(ctx, zone.ID, e.CatchallPrivate, privateDns); err != nil {
		return fmt.Errorf("Couldn't create private catchall dns entry: %s %s - %v", e.CatchallPrivate, privateDns, err)
	}
	if err := cli.createRecord(ctx, zone.ID, e.Private, privateDns); err != nil {
		return fmt.Errorf("Couldn't create private dns entry: %s %s - %v", e.Private, privateDns, err)
	}
	if err := cli.createRecord(ctx, zone.ID, e.Fleet, instances[0].PublicDNSName); err != nil {
		return fmt.Errorf("Couldn't create fleet dns entry: %s %s - %v", e.Fleet, instances[0].PublicDNSName, err)
	}
	return nil
}

// DeleteEntries deletes DNS entries, given a stack name, and list of Entries to delete.
func (cli *CloudFlareDNS) DeleteEntries(name string, e *Entries) error {
	ctx := context.TODO()

	client := cli.client()

	// First we need to find our zone
	zone, err := cli.findZone(ctx, e.Zone)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	records, err := client.Records.List(ctx, zone.ID)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	for _, record := range records {
		entriesRecord := record.Name == e.Catchall ||
			record.Name == e.CatchallPrivate ||
			record.Name == e.Public ||
			record.Name == e.Private ||
			record.Name == e.Fleet
		if entriesRecord {
			if err := client.Records.Delete(ctx, zone.ID, record.ID); err != nil {
				return errgo.Mask(err, errgo.Any)
			}
		}
	}
	return nil
}

// Update updates DNS records, given a swarm name, CNAME, dns content, and Entries.
func (cli *CloudFlareDNS) Update(swarmName, cname, dns string, e *Entries) error {
	ctx := context.TODO()
	client := cli.client()

	zone, err := cli.findZone(ctx, e.Zone)
	if err != nil {
		return errgo.Mask(err, errgo.Any)
	}

	client.Records.Patch(ctx, &cloudflare.Record{
		ZoneID:  zone.ID,
		Type:    "CNAME",
		Name:    cname,
		Content: dns,
	})

	return nil
}

func (api *CloudFlareDNS) client() *cloudflare.Client {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api._client != nil {
		return api._client
	}

	options := cloudflare.Options{
		Email: os.Getenv("CLOUDFLARE_EMAIL"),
		Key:   os.Getenv("CLOUDFLARE_TOKEN"),
	}
	if options.Email == "" || options.Key == "" {
		panic("environment variables CLOUDFLARE_EMAIL or CLOUDFLARE_TOKEN missing")
	}
	api._client = cloudflare.New(&options)

	return api._client
}

func (cli *CloudFlareDNS) createRecord(ctx context.Context, zoneID, cname, dns string) error {
	err := cli.client().Records.Create(ctx, &cloudflare.Record{
		ZoneID:  zoneID,
		Type:    "CNAME",
		Name:    cname,
		Content: dns,
	})
	return errgo.Mask(err, errgo.Any)
}

func (cli *CloudFlareDNS) findZone(ctx context.Context, domain string) (*cloudflare.Zone, error) {
	client := cli.client()
	zones, err := client.Zones.List(ctx)
	if err != nil {
		return nil, errgo.Mask(err, errgo.Any)
	}

	for _, z := range zones {
		if z.Name == domain {
			return z, nil
		}
	}
	return nil, errgo.Newf("no zone for domain %s found", domain)
}
