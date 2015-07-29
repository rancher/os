package docker

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

type configEnvironment struct {
	cfg *config.CloudConfig
}

func appendEnv(array []string, key, value string) []string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		key = parts[1]
	}

	return append(array, fmt.Sprintf("%s=%s", key, value))
}

func lookupKeys(cfg *config.CloudConfig, keys ...string) []string {
	for _, key := range keys {
		if strings.HasSuffix(key, "*") {
			result := []string{}
			for envKey, envValue := range cfg.Rancher.Environment {
				keyPrefix := key[:len(key)-1]
				if strings.HasPrefix(envKey, keyPrefix) {
					result = appendEnv(result, envKey, envValue)
				}
			}

			if len(result) > 0 {
				return result
			}
		} else if value, ok := cfg.Rancher.Environment[key]; ok {
			return appendEnv([]string{}, key, value)
		}
	}

	return []string{}
}

func (c *configEnvironment) Lookup(key, serviceName string, serviceConfig *project.ServiceConfig) []string {
	fullKey := fmt.Sprintf("%s/%s", serviceName, key)
	return lookupKeys(c.cfg, fullKey, key)
}

func RunServices(name string, cfg *config.CloudConfig, configs map[string]*project.ServiceConfig) error {
	network := false
	projectEvents := make(chan project.ProjectEvent)
	p := project.NewProject(name, NewContainerFactory(cfg))
	p.EnvironmentLookup = &configEnvironment{cfg: cfg}
	p.AddListener(projectEvents)
	enabled := make(map[string]bool)

	for name, serviceConfig := range configs {
		if err := p.AddConfig(name, serviceConfig); err != nil {
			log.Infof("Failed loading service %s", name)
			continue
		}
		enabled[name] = true
	}

	p.ReloadCallback = func() error {
		if p.Name != "system-init" {
			return nil
		}

		if err := cfg.Reload(); err != nil {
			return err
		}

		for service, serviceEnabled := range cfg.Rancher.ServicesInclude {
			if !serviceEnabled {
				continue
			}

			if en, ok := enabled[service]; ok && en {
				continue
			}

			bytes, err := LoadServiceResource(service, network, cfg)
			if err != nil {
				if err == util.ErrNoNetwork {
					log.Debugf("Can not load %s, networking not enabled", service)
				} else {
					log.Errorf("Failed to load %s : %v", service, err)
				}
				continue
			}

			if err := p.Load(bytes); err != nil {
				log.Errorf("Failed to load %s : %v", service, err)
				continue
			}

			enabled[service] = true
		}

		for service, config := range cfg.Rancher.Services {
			if en, ok := enabled[service]; ok && en {
				continue
			}

			if err := p.AddConfig(service, config); err != nil {
				log.Errorf("Failed to load %s : %v", service, err)
				continue
			}
			enabled[service] = true
		}

		return nil
	}

	go func() {
		for event := range projectEvents {
			if event.Event == project.CONTAINER_STARTED && event.ServiceName == "network" {
				network = true
			}
		}
	}()

	if err := p.ReloadCallback(); err != nil {
		log.Errorf("Failed to reload %s : %v", name, err)
		return err
	}
	return p.Up()
}

func LoadServiceResource(name string, network bool, cfg *config.CloudConfig) ([]byte, error) {
	return util.LoadResource(name, network, cfg.Rancher.Repositories.ToArray())
}
