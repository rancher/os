package docker

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/project"
)

type configEnvironemnt struct {
	cfg *config.Config
}

func appendEnv(array []string, key, value string) []string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		key = parts[1]
	}

	return append(array, fmt.Sprintf("%s=%s", key, value))
}

func lookupKeys(cfg *config.Config, keys ...string) []string {
	for _, key := range keys {
		if strings.HasSuffix(key, "*") {
			result := []string{}
			for envKey, envValue := range cfg.Environment {
				keyPrefix := key[:len(key)-1]
				if strings.HasPrefix(envKey, keyPrefix) {
					result = appendEnv(result, envKey, envValue)
				}
			}

			if len(result) > 0 {
				return result
			}
		} else if value, ok := cfg.Environment[key]; ok {
			return appendEnv([]string{}, key, value)
		}
	}

	return []string{}
}

func (c *configEnvironemnt) Lookup(key, serviceName string, serviceConfig *project.ServiceConfig) []string {
	fullKey := fmt.Sprintf("%s/%s", serviceName, key)
	return lookupKeys(c.cfg, fullKey, key)
}

func RunServices(name string, cfg *config.Config, configs map[string]*project.ServiceConfig) error {
	network := false
	projectEvents := make(chan project.ProjectEvent)
	p := project.NewProject(name, NewContainerFactory(cfg))
	p.EnvironmentLookup = &configEnvironemnt{cfg: cfg}
	p.AddListener(projectEvents)
	enabled := make(map[string]bool)

	for name, serviceConfig := range configs {
		if err := p.AddConfig(name, serviceConfig); err != nil {
			log.Infof("Failed loading service %s", name)
		}
	}

	p.ReloadCallback = func() error {
		err := cfg.Reload()
		if err != nil {
			return err
		}

		for service, serviceEnabled := range cfg.ServicesInclude {
			if !serviceEnabled {
				continue
			}

			if _, ok := enabled[service]; ok {
				continue
			}

			//if config, ok := cfg.BundledServices[service]; ok {
			//	for name, s := range config.SystemContainers {
			//		if err := p.AddConfig(name, s); err != nil {
			//			log.Errorf("Failed to load %s : %v", name, err)
			//		}
			//	}
			//} else {
			bytes, err := LoadServiceResource(service, network, cfg)
			if err != nil {
				if err == util.ErrNoNetwork {
					log.Debugf("Can not load %s, networking not enabled", service)
				} else {
					log.Errorf("Failed to load %s : %v", service, err)
				}
				continue
			}

			err = p.Load(bytes)
			if err != nil {
				log.Errorf("Failed to load %s : %v", service, err)
				continue
			}
			//}

			enabled[service] = true
		}

		for service, config := range cfg.Services {
			if _, ok := enabled[service]; ok {
				continue
			}

			err = p.AddConfig(service, config)
			if err != nil {
				log.Errorf("Failed to load %s : %v", service, err)
				continue
			}

			enabled[service] = true
		}

		return nil
	}

	go func() {
		for event := range projectEvents {
			if event.Event == project.CONTAINER_STARTED && event.Service.Name() == "network" {
				network = true
			}
		}
	}()

	err := p.ReloadCallback()
	if err != nil {
		log.Errorf("Failed to reload %s : %v", name, err)
		return err
	}
	return p.Up()
}

func LoadServiceResource(name string, network bool, cfg *config.Config) ([]byte, error) {
	return util.LoadResource(name, network, cfg.Repositories.ToArray())
}
