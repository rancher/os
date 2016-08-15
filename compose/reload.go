package compose

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/util/network"
)

func LoadService(p *project.Project, cfg *config.CloudConfig, useNetwork bool, service string) error {
	bytes, err := network.LoadServiceResource(service, useNetwork, cfg)
	if err != nil {
		return err
	}

	m := map[interface{}]interface{}{}
	if err = yaml.Unmarshal(bytes, &m); err != nil {
		return fmt.Errorf("Failed to parse YAML configuration for %s: %v", service, err)
	}

	m = adjustContainerNames(m)

	bytes, err = yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("Failed to marshal YAML configuration for %s: %v", service, err)
	}

	if err = p.Load(bytes); err != nil {
		return fmt.Errorf("Failed to load %s: %v", service, err)
	}

	return nil
}

func projectReload(p *project.Project, useNetwork *bool, loadConsole bool, environmentLookup *docker.ConfigEnvironment, authLookup *docker.ConfigAuthLookup) func() error {
	enabled := map[interface{}]interface{}{}
	return func() error {
		cfg := config.LoadConfig()

		environmentLookup.SetConfig(cfg)
		authLookup.SetConfig(cfg)

		enabled = addServices(p, enabled, cfg.Rancher.Services)

		for service, serviceEnabled := range cfg.Rancher.ServicesInclude {
			if _, ok := enabled[service]; ok || !serviceEnabled {
				continue
			}

			if err := LoadService(p, cfg, *useNetwork, service); err != nil {
				if err != network.ErrNoNetwork {
					log.Error(err)
				}
				continue
			}

			enabled[service] = service
		}

		if loadConsole && cfg.Rancher.Console != "" && cfg.Rancher.Console != "default" {
			if err := LoadService(p, cfg, *useNetwork, cfg.Rancher.Console); err != nil && err != network.ErrNoNetwork {
				log.Error(err)
			}
		}

		if cfg.Rancher.Docker.Engine != "" {
			p.ServiceConfigs.Add("docker", &composeConfig.ServiceConfig{})
			if err := LoadService(p, cfg, *useNetwork, cfg.Rancher.Docker.Engine); err != nil && err != network.ErrNoNetwork {
				log.Error(err)
			}
		}

		return nil
	}
}
