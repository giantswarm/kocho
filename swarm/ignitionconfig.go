package swarm

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/giantswarm/kocho/swarm/types"

	"github.com/juju/errgo"
)

type primaryIgnitionConfig struct {
	DiscoveryUrl  string
	YochuVersion  string
	Tags          string
	EtcdVersion   string
	FleetVersion  string
	DockerVersion string
	K8sVersion    string
	RktVersion    string
}

type secondaryIgnitionConfig struct {
	EtcdPeers        string
	YochuVersion     string
	Tags             string
	EtcdVersion      string
	FleetVersion     string
	DockerVersion    string
	EtcdDiscoveryURL string
	K8sVersion       string
	RktVersion       string
}

// ClusterBootstrap
const (
	primaryIgnitionConfigTemplateName    = "/ignition/primary-ignition.tmpl"
	secondaryIgnitionConfigTemplateName  = "/ignition/secondary-ignition.tmpl"
	standaloneIgnitionConfigTemplateName = "/ignition/standalone-ignition.tmpl"
)

func createIgnitionConfig(flags swarmtypes.CreateFlags) (string, error) {
	// add default tags for the primary instances
	tags := fmt.Sprintf("role=%s,%s", flags.Type, flags.Tags)

	switch flags.Type {
	case "primary":
		return createPrimaryIgnitionConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	case "standalone":
		return createStandaloneIgnitionConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	case "secondary":
		if flags.EtcdPeers == "" {
			return "", errors.New("etcd peers for secondary ignition config are missing")
		}
		if flags.EtcdDiscoveryURL == "" {
			return "", errors.New("etcd discovery url for secondary ignition config are missing")
		}
		if !strings.HasPrefix(flags.EtcdPeers, "http") {
			return "", errors.New("etcd peers have to start with http/https protocol definition")
		}
		return createSecondaryIgnitionConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.EtcdPeers, flags.EtcdDiscoveryURL, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	}

	return "", errgo.New(fmt.Sprintf("type not valid: %s", flags.Type))
}

func createPrimaryIgnitionConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, k8sVersion, rktVersion, templateDir, tags string) (string, error) {
	discoveryUrl, err := getNewDiscoveryUrl()
	if err != nil {
		return "", errgo.Mask(err)
	}

	ignitionConfigTemplatePath := path.Join(templateDir, primaryIgnitionConfigTemplateName)

	return parseIgnitionConfigTemplate(ignitionConfigTemplatePath, primaryIgnitionConfig{
		DiscoveryUrl:  discoveryUrl,
		YochuVersion:  yochuVersion,
		Tags:          tags,
		EtcdVersion:   etcdVersion,
		FleetVersion:  fleetVersion,
		DockerVersion: dockerVersion,
		K8sVersion:    k8sVersion,
		RktVersion:    rktVersion,
	})
}

func createStandaloneIgnitionConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	discoveryUrl, err := getNewDiscoveryUrl()
	if err != nil {
		return "", errgo.Mask(err)
	}

	ignitionConfigTemplatePath := path.Join(templateDir, standaloneIgnitionConfigTemplateName)

	return parseIgnitionConfigTemplate(ignitionConfigTemplatePath, primaryIgnitionConfig{
		DiscoveryUrl:  discoveryUrl,
		YochuVersion:  yochuVersion,
		Tags:          tags,
		FleetVersion:  fleetVersion,
		EtcdVersion:   etcdVersion,
		DockerVersion: dockerVersion,
		K8sVersion:    k8sVersion,
		RktVersion:    rktVersion,
	})
}

func createSecondaryIgnitionConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, etcdPeers, etcdDiscoveryURL, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	ignitionConfigTemplatePath := path.Join(templateDir, secondaryIgnitionConfigTemplateName)

	return parseIgnitionConfigTemplate(ignitionConfigTemplatePath, secondaryIgnitionConfig{
		YochuVersion:     yochuVersion,
		Tags:             tags,
		EtcdVersion:      etcdVersion,
		FleetVersion:     fleetVersion,
		DockerVersion:    dockerVersion,
		EtcdPeers:        etcdPeers,
		EtcdDiscoveryURL: etcdDiscoveryURL,
		K8sVersion:       k8sVersion,
		RktVersion:       rktVersion,
	})
}

func parseIgnitionConfigTemplate(ignitionConfigTemplatePath string, cfg interface{}) (string, error) {
	absoluteIgnitionConfigTemplatePath, err := filepath.Abs(ignitionConfigTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	templateData, err := ioutil.ReadFile(absoluteIgnitionConfigTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	var tmpl *template.Template
	tmpl, err = template.New("ignition-config").Parse(string(templateData))
	if err != nil {
		return "", errgo.Mask(err)
	}

	buffer := new(bytes.Buffer)
	if err := tmpl.Execute(buffer, cfg); err != nil {
		return "", errgo.Mask(err)
	}

	return string(buffer.Bytes()), nil
}
