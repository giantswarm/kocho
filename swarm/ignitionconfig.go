package swarm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/giantswarm/kocho/swarm/types"

	"github.com/coreos/ignition/config/v1/types"
	"github.com/juju/errgo"
	"gopkg.in/yaml.v2"
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
	primaryIgnitionConfigTemplateName    = "primary-ignition.tmpl"
	secondaryIgnitionConfigTemplateName  = "secondary-ignition.tmpl"
	standaloneIgnitionConfigTemplateName = "standalone-ignition.tmpl"
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

	ignitionTemplate, err := parseIgnitionConfigTemplate(ignitionConfigTemplatePath, primaryIgnitionConfig{
		DiscoveryUrl:  discoveryUrl,
		YochuVersion:  yochuVersion,
		Tags:          tags,
		EtcdVersion:   etcdVersion,
		FleetVersion:  fleetVersion,
		DockerVersion: dockerVersion,
		K8sVersion:    k8sVersion,
		RktVersion:    rktVersion,
	})
	if err != nil {
		return "", err
	}

	ignitionJSON, err := convertTemplatetoJSON([]byte(ignitionTemplate), true)
	if err != nil {
		return "", err
	}

	return string(ignitionJSON[:]), nil
}

func createStandaloneIgnitionConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	discoveryUrl, err := getNewDiscoveryUrl()
	if err != nil {
		return "", errgo.Mask(err)
	}

	ignitionConfigTemplatePath := path.Join(templateDir, standaloneIgnitionConfigTemplateName)

	ignitionTemplate, err := parseIgnitionConfigTemplate(ignitionConfigTemplatePath, primaryIgnitionConfig{
		DiscoveryUrl:  discoveryUrl,
		YochuVersion:  yochuVersion,
		Tags:          tags,
		FleetVersion:  fleetVersion,
		EtcdVersion:   etcdVersion,
		DockerVersion: dockerVersion,
		K8sVersion:    k8sVersion,
		RktVersion:    rktVersion,
	})
	if err != nil {
		return "", err
	}

	ignitionJSON, err := convertTemplatetoJSON([]byte(ignitionTemplate), true)
	if err != nil {
		return "", err
	}

	return string(ignitionJSON[:]), nil
}

func createSecondaryIgnitionConfig(yochuVersion, fleetVersion, etcdVersion, dockerVersion, etcdPeers, etcdDiscoveryURL, k8sVersion, rktVersion, templateDir string, tags string) (string, error) {
	ignitionConfigTemplatePath := path.Join(templateDir, secondaryIgnitionConfigTemplateName)

	ignitionTemplate, err := parseIgnitionConfigTemplate(ignitionConfigTemplatePath, secondaryIgnitionConfig{
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
	if err != nil {
		return "", err
	}

	ignitionJSON, err := convertTemplatetoJSON([]byte(ignitionTemplate), true)
	if err != nil {
		return "", err
	}

	return string(ignitionJSON[:]), nil
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

func convertTemplatetoJSON(dataIn []byte, pretty bool) ([]byte, error) {
	cfg := types.Config{}

	if err := yaml.Unmarshal(dataIn, &cfg); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to unmarshal input: %v", err))
	}

	var inCfg interface{}
	if err := yaml.Unmarshal(dataIn, &inCfg); err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to unmarshal input: %v", err))
	}

	if hasUnrecognizedKeys(inCfg, reflect.TypeOf(cfg)) {
		return nil, errors.New(fmt.Sprintf("Unrecognized keys in input, aborting."))
	}

	var (
		dataOut []byte
		err     error
	)

	if pretty {
		dataOut, err = json.MarshalIndent(&cfg, "", "  ")
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to marshal output: %v", err))
		}
		dataOut = append(dataOut, '\n')
	} else {
		dataOut, err = json.Marshal(&cfg)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to marshal output: %v", err))
		}
	}

	return dataOut, nil
}

func hasUnrecognizedKeys(inCfg interface{}, refType reflect.Type) (warnings bool) {
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}
	switch inCfg.(type) {
	case map[interface{}]interface{}:
		ks := inCfg.(map[interface{}]interface{})
	keys:
		for key := range ks {
			for i := 0; i < refType.NumField(); i++ {
				sf := refType.Field(i)
				tv := sf.Tag.Get("yaml")
				if tv == key {
					if warn := hasUnrecognizedKeys(ks[key], sf.Type); warn {
						warnings = true
					}
					continue keys
				}
			}

			fmt.Println("Unrecognized ignition property: %v", key)
			warnings = true
		}
	case []interface{}:
		ks := inCfg.([]interface{})
		for i := range ks {
			if warn := hasUnrecognizedKeys(ks[i], refType.Elem()); warn {
				warnings = true
			}
		}
	default:
	}
	return
}
