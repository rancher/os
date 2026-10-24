package compose

import (
	"fmt"
	"strings"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/docker"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"
	"github.com/burmilla/os/pkg/util/network"
	"github.com/burmilla/os/pkg/util/versions"

	composeConfig "github.com/burmilla/os/pkg/libcompose/config"
	"github.com/burmilla/os/pkg/libcompose/project"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

func LoadService(p *project.Project, cfg *config.CloudConfig, useNetwork bool, service string) error {
	// First check the multi engine service file.
	// If the name has been found in multi enging service file and matches, will not execute network.LoadServiceResource
	// Otherwise will execute network.LoadServiceResource
	bytes, err := network.LoadMultiEngineResource(service)
	if err != nil || bytes == nil {
		bytes, err = network.LoadServiceResource(service, useNetwork, cfg)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	m := map[interface{}]interface{}{}
	if err = yaml.Unmarshal(bytes, &m); err != nil {
		e := fmt.Errorf("Failed to parse YAML configuration for %s: %v", service, err)
		log.Error(e)
		return e
	}

	m = adjustContainerNames(m)

	bytes, err = yaml.Marshal(m)
	if err != nil {
		e := fmt.Errorf("Failed to marshal YAML configuration for %s: %v", service, err)
		log.Error(e)
		return e
	}

	if err = p.Load(bytes); err != nil {
		e := fmt.Errorf("Failed to load %s: %v", service, err)
		log.Error(e)
		return e
	}

	return nil
}

func LoadSpecialService(p *project.Project, cfg *config.CloudConfig, serviceName, serviceValue string) error {
	if serviceValue == "" {
		return fmt.Errorf("empty service value for %q", serviceName)
	}

	// Save config in case load fails so we can roll back
	previousConfig, ok := p.ServiceConfigs.Get(serviceName)

	p.ServiceConfigs.Add(serviceName, &composeConfig.ServiceConfig{})

	if err := LoadService(p, cfg, true, serviceValue); err != nil {
		// Rollback to previous config to avoid leaving an empty placeholder
		if ok {
			p.ServiceConfigs.Add(serviceName, previousConfig)
		}
		return err
	}

	return nil
}

func loadConsoleService(cfg *config.CloudConfig, p *project.Project) error {
	if cfg.Rancher.Console == "" || cfg.Rancher.Console == "default" {
		return nil
	}
	return LoadSpecialService(p, cfg, "console", cfg.Rancher.Console)
}

func loadEngineService(cfg *config.CloudConfig, p *project.Project) error {
	engine := cfg.Rancher.Docker.Engine
	if engine == "" {
		engine = cfg.Rancher.Defaults.Docker.Engine
	}
	if engine == "" || (engine != "latest" && engine == cfg.Rancher.Defaults.Docker.Engine) {
		return nil
	}

	// The configured engine may be missing from the service repository of
	// this OS version (e.g. an engine selected on an older release). Fall
	// back to the default engine instead of leaving the system without Docker.
	// "latest" is not an engine in the repository (there is no os-docker:latest
	// image), so it always resolves to the latest available engine.
	engines, err := network.GetEngines(cfg.Rancher.Repositories.ToArray())
	if err == nil {
		if fallback, ok := resolveEngine(engine, cfg.Rancher.Defaults.Docker.Engine, engines); !ok {
			if engine != "latest" {
				log.Warnf("Docker engine %s is not available in the service repositories, falling back to %s. Use 'ros engine list' and 'ros engine switch' to select an available engine.", engine, fallback)
			}
			engine = fallback
		}
	}
	if engine == "latest" {
		// Engine index could not be read, keep the built-in docker service.
		return nil
	}

	return LoadSpecialService(p, cfg, "docker", engine)
}

// resolveEngine returns engine and true if it can be loaded: it is a local
// file or URL, it is listed in available, or the engine index could not be
// read (available is empty). Otherwise it returns the engine to fall back to
// and false: defaultEngine if it is listed in available, else the latest
// available engine, as with 'ros engine switch latest'.
func resolveEngine(engine, defaultEngine string, available []string) (string, bool) {
	if len(available) == 0 || isLocalOrURL(engine) || util.Contains(available, engine) {
		return engine, true
	}
	if util.Contains(available, defaultEngine) {
		return defaultEngine, false
	}
	sorted := append([]string(nil), available...)
	versions.SortEngines(sorted)
	return sorted[len(sorted)-1], false
}

func isLocalOrURL(service string) bool {
	return strings.HasPrefix(service, "/") || strings.HasPrefix(service, "http:/") || strings.HasPrefix(service, "https:/")
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
					log.Errorf("Failed to load service(%s): %v", service, err)
				}
				continue
			}

			enabled[service] = service
		}

		if !*useNetwork {
			return nil
		}

		if loadConsole {
			if err := loadConsoleService(cfg, p); err != nil {
				log.Errorf("Failed to load rancher.console=(%s): %v", cfg.Rancher.Console, err)
			}
		}

		if err := loadEngineService(cfg, p); err != nil {
			log.Errorf("Failed to load rancher.docker.engine=(%s): %v", cfg.Rancher.Docker.Engine, err)
		}

		return nil
	}
}
