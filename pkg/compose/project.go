package compose

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/burmilla/os/config"
	rosDocker "github.com/burmilla/os/pkg/docker"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"
	"github.com/burmilla/os/pkg/util/network"

	"context"
	"github.com/burmilla/os/pkg/libcompose/cli/logger"
	composeConfig "github.com/burmilla/os/pkg/libcompose/config"
	"github.com/burmilla/os/pkg/libcompose/docker"
	composeClient "github.com/burmilla/os/pkg/libcompose/docker/client"
	"github.com/burmilla/os/pkg/libcompose/project"
	"github.com/burmilla/os/pkg/libcompose/project/events"
	"github.com/burmilla/os/pkg/libcompose/project/options"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	composetypes "github.com/compose-spec/compose-go/types"
	dockerClient "github.com/docker/engine-api/client"
)

// CreateService creates a single service from a compose-go ServiceConfig.
func CreateService(cfg *config.CloudConfig, name string, serviceConfig *composetypes.ServiceConfig) (project.Service, error) {
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	p, err := newProject("once", cfg, nil, nil)
	if err != nil {
		return nil, err
	}

	serviceConfig.Name = name
	if err := p.AddConfig(name, serviceConfig); err != nil {
		return nil, err
	}

	return p.CreateService(name)
}

// CreateServiceSet creates a project with the given v1 service configs.
func CreateServiceSet(name string, cfg *config.CloudConfig, configs map[string]*composeConfig.ServiceConfigV1) (*project.Project, error) {
	p, err := newProject(name, cfg, nil, nil)
	if err != nil {
		return nil, err
	}

	addServices(p, map[interface{}]interface{}{}, configs)

	return p, nil
}

// RunServiceSet creates and starts a set of services from v1 configs.
func RunServiceSet(name string, cfg *config.CloudConfig, configs map[string]*composeConfig.ServiceConfigV1) (*project.Project, error) {
	p, err := CreateServiceSet(name, cfg, configs)
	if err != nil {
		return nil, err
	}
	return p, p.Up(context.Background(), options.Up{
		Log: cfg.Rancher.Log,
	})
}

func GetProject(cfg *config.CloudConfig, networkingAvailable, loadConsole bool) (*project.Project, error) {
	return newCoreServiceProject(cfg, networkingAvailable, loadConsole)
}

func newProject(name string, cfg *config.CloudConfig, environmentLookup composeConfig.EnvironmentLookup, authLookup *rosDocker.ConfigAuthLookup) (*project.Project, error) {
	clientFactory, err := rosDocker.NewClientFactory(composeClient.Options{})
	if err != nil {
		return nil, err
	}

	if environmentLookup == nil {
		environmentLookup = rosDocker.NewConfigEnvironment(cfg)
	}
	if authLookup == nil {
		authLookup = rosDocker.NewConfigAuthLookup(cfg)
	}

	serviceFactory := &rosDocker.ServiceFactory{
		Deps: map[string][]string{},
	}
	context := &docker.Context{
		ClientFactory: clientFactory,
		AuthLookup:    authLookup,
		Context: project.Context{
			ProjectName:       name,
			EnvironmentLookup: environmentLookup,
			ServiceFactory:    serviceFactory,
			LoggerFactory:     logger.NewColorLoggerFactory(),
		},
	}
	serviceFactory.Context = context

	authLookup.SetContext(context)

	return docker.NewProject(context, &composeConfig.ParseOptions{
		Interpolate: true,
		Validate:    false,
		Preprocess:  preprocessServiceMap,
	})
}

// preprocessServiceMap converts all values in "environment" and "labels" keys
// to strings, which is required for proper variable interpolation. Other keys
// are left as-is to preserve their original types (int, bool, etc.).
func preprocessServiceMap(serviceMap composeConfig.RawServiceMap) (composeConfig.RawServiceMap, error) {
	newServiceMap := make(composeConfig.RawServiceMap)

	for serviceName, service := range serviceMap {
		newServiceMap[serviceName] = make(composeConfig.RawService)
		for key, value := range service {
			stringifyValues := key == "environment" || key == "labels"
			newServiceMap[serviceName][key] = preprocess(value, stringifyValues)
		}
	}

	return newServiceMap, nil
}

func preprocess(item interface{}, replaceTypes bool) interface{} {
	switch typedDatas := item.(type) {

	case map[interface{}]interface{}:
		newMap := make(map[interface{}]interface{})

		for key, value := range typedDatas {
			newMap[key] = preprocess(value, replaceTypes)
		}
		return newMap

	case []interface{}:
		// newArray := make([]interface{}, 0) will cause golint to complain
		var newArray []interface{}
		newArray = make([]interface{}, 0)

		for _, value := range typedDatas {
			newArray = append(newArray, preprocess(value, replaceTypes))
		}
		return newArray

	default:
		if replaceTypes {
			return fmt.Sprint(item)
		}
		return item
	}
}

// addServices converts v1 service configs and adds them to the project.
// It uses content hashing to skip services whose config has not changed,
// avoiding unnecessary service recreation during reloads.
func addServices(p *project.Project, enabled map[interface{}]interface{}, configs map[string]*composeConfig.ServiceConfigV1) map[interface{}]interface{} {
	serviceConfigsV2, _ := composeConfig.ConvertServices(configs)

	mapCopied := false
	for name, serviceConfig := range serviceConfigsV2 {
		hash := composeConfig.GetServiceHash(name, serviceConfig)

		// Skip if this service's config hash hasn't changed since last load
		if enabled[name] == hash {
			continue
		}

		if err := p.AddConfig(name, serviceConfig); err != nil {
			log.Infof("Failed loading service %s", name)
			continue
		}

		// Copy-on-write: only copy the map when we first need to modify it
		if !mapCopied {
			enabled = util.MapCopy(enabled)
			mapCopied = true
		}
		enabled[name] = hash
	}
	return enabled
}

func adjustContainerNames(m map[interface{}]interface{}) map[interface{}]interface{} {
	for k, v := range m {
		if k, ok := k.(string); ok {
			if v, ok := v.(map[interface{}]interface{}); ok {
				if _, ok := v["container_name"]; !ok {
					v["container_name"] = k
				}
			}
		}
	}
	return m
}

func newCoreServiceProject(cfg *config.CloudConfig, useNetwork, loadConsole bool) (*project.Project, error) {
	environmentLookup := rosDocker.NewConfigEnvironment(cfg)
	authLookup := rosDocker.NewConfigAuthLookup(cfg)

	p, err := newProject("os", cfg, environmentLookup, authLookup)
	if err != nil {
		return nil, err
	}

	projectEvents := make(chan events.Event)
	p.AddListener(project.NewDefaultListener(p))
	p.AddListener(projectEvents)

	p.ReloadCallback = projectReload(p, &useNetwork, loadConsole, environmentLookup, authLookup)

	go func() {
		for event := range projectEvents {
			if event.EventType == events.ContainerStarted && event.ServiceName == "network" {
				useNetwork = true
			}
		}
	}()

	err = p.ReloadCallback()
	if err != nil {
		log.Errorf("Failed to reload os: %v", err)
		return nil, err
	}

	return p, nil
}

func StageServices(cfg *config.CloudConfig, services ...string) error {
	p, err := newProject("stage-services", cfg, nil, nil)
	if err != nil {
		return err
	}

	// read engine services
	composeConfigs := map[string]composeConfig.ServiceConfigV1{}
	if _, err := os.Stat(config.MultiDockerConfFile); err == nil {
		// read from engine compose
		multiEngineBytes, err := ioutil.ReadFile(config.MultiDockerConfFile)
		if err != nil {
			return fmt.Errorf("Failed to read %s : %v", config.MultiDockerConfFile, err)
		}
		err = yaml.Unmarshal(multiEngineBytes, &composeConfigs)
		if err != nil {
			return fmt.Errorf("Failed to unmarshal %s : %v", config.MultiDockerConfFile, err)
		}
	}

	for _, service := range services {
		var bytes []byte
		foundServiceConfig := map[string]composeConfig.ServiceConfigV1{}

		if _, ok := composeConfigs[service]; ok {
			foundServiceConfig[service] = composeConfigs[service]
			bytes, err = yaml.Marshal(foundServiceConfig)
		} else {
			bytes, err = network.LoadServiceResource(service, true, cfg)
		}

		if err != nil {
			return fmt.Errorf("Failed to load %s : %v", service, err)
		}

		m := map[interface{}]interface{}{}
		if err := yaml.Unmarshal(bytes, &m); err != nil {
			return fmt.Errorf("Failed to parse YAML configuration: %s : %v", service, err)
		}

		bytes, err = yaml.Marshal(m)
		if err != nil {
			return fmt.Errorf("Failed to marshal YAML configuration: %s : %v", service, err)
		}

		err = p.Load(bytes)
		if err != nil {
			return fmt.Errorf("Failed to load %s : %v", service, err)
		}
	}

	// Reduce service configurations to just image and labels
	needToPull := false
	var client, userClient, systemClient dockerClient.APIClient
	for _, serviceName := range p.ServiceConfigs.Keys() {
		serviceConfig, _ := p.ServiceConfigs.Get(serviceName)

		// test to see if we need to Pull
		if serviceConfig.Labels[config.ScopeLabel] != config.System {
			if userClient == nil {
				userClient, err = rosDocker.NewDefaultClient()
				if err != nil {
					log.Error(err)
				}

			}
			client = userClient
		} else {
			if systemClient == nil {
				systemClient, err = rosDocker.NewSystemClient()
				if err != nil {
					log.Error(err)
				}
				client = systemClient
			}
		}
		if client != nil {
			_, _, err := client.ImageInspectWithRaw(context.Background(), serviceConfig.Image, false)
			if err == nil {
				log.Infof("Service %s using local image %s", serviceName, serviceConfig.Image)
				continue
			}
		}
		needToPull = true

		p.ServiceConfigs.Add(serviceName, &composetypes.ServiceConfig{
			Image:  serviceConfig.Image,
			Labels: serviceConfig.Labels,
		})
	}

	if needToPull {
		return p.Pull(context.Background())
	}
	return nil
}
