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

type primaryCloudConfig struct {
	DiscoveryUrl  string
	YochuVersion  string
	Tags          string
	EtcdVersion   string
	FleetVersion  string
	DockerVersion string
	K8sVersion    string
	RktVersion    string
}

type secondaryCloudConfig struct {
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
	discoveryService                  = "https://discovery.etcd.io/new"
	primaryCloudConfigTemplateName    = "primary-cloudconfig.tmpl"
	secondaryCloudConfigTemplateName  = "secondary-cloudconfig.tmpl"
	standaloneCloudConfigTemplateName = "standalone-cloudconfig.tmpl"
)

func createCloudConfig(flags swarmtypes.CreateFlags) (string, error) {
	// add default tags for the primary instances
	tags := fmt.Sprintf("role=%s,%s", flags.Type, flags.Tags)

	switch flags.Type {
	case "primary":
		return createPrimaryCloudConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	case "standalone":
		return createStandaloneCloudConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	case "secondary":
		if flags.EtcdPeers == "" {
			return "", errors.New("etcd peers for secondary cloud-config are missing")
		}
		if flags.EtcdDiscoveryURL == "" {
			return "", errors.New("etcd discovery url for secondary cloud-config are missing")
		}
		if !strings.HasPrefix(flags.EtcdPeers, "http") {
			return "", errors.New("etcd peers have to start with http/https protocol definition")
		}
		return createSecondaryCloudConfig(flags.YochuVersion, flags.FleetVersion, flags.EtcdVersion, flags.DockerVersion, flags.EtcdPeers, flags.EtcdDiscoveryURL, flags.K8sVersion, flags.RktVersion, flags.TemplateDir, tags)
	}

	return "", errgo.New(fmt.Sprintf("type not valid: %s", flags.Type))
}

func createPrimaryCloudConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, k8sVersion, rktVersion, templateDir, tags string) (string, error) {
	discoveryUrl, err := getNewDiscoveryUrl()
	if err != nil {
		return "", errgo.Mask(err)
	}

	cloudConfigTemplatePath := path.Join(templateDir, primaryCloudConfigTemplateName)

	return parseCloudConfigTemplate(cloudConfigTemplatePath, primaryCloudConfig{
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

func createStandaloneCloudConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	discoveryUrl, err := getNewDiscoveryUrl()
	if err != nil {
		return "", errgo.Mask(err)
	}

	cloudConfigTemplatePath := path.Join(templateDir, standaloneCloudConfigTemplateName)

	return parseCloudConfigTemplate(cloudConfigTemplatePath, primaryCloudConfig{
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

func createSecondaryCloudConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, etcdPeers, etcdDiscoveryURL, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	cloudConfigTemplatePath := path.Join(templateDir, secondaryCloudConfigTemplateName)

	return parseCloudConfigTemplate(cloudConfigTemplatePath, secondaryCloudConfig{
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

func parseCloudConfigTemplate(cloudConfigTemplatePath string, cfg interface{}) (string, error) {
	absoluteCloudConfigTemplatePath, err := filepath.Abs(cloudConfigTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	templateData, err := ioutil.ReadFile(absoluteCloudConfigTemplatePath)
	if err != nil {
		return "", errgo.Mask(err)
	}

	var tmpl *template.Template
	tmpl, err = template.New("cloud-config").Parse(string(templateData))
	if err != nil {
		return "", errgo.Mask(err)
	}

	buffer := new(bytes.Buffer)
	if err := tmpl.Execute(buffer, cfg); err != nil {
		return "", errgo.Mask(err)
	}

	return string(buffer.Bytes()), nil
}
