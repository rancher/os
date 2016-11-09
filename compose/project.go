package compose

import (
	"fmt"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/docker/libcompose/cli/logger"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	composeClient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/events"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/config"
	rosDocker "github.com/rancher/os/docker"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

func CreateService(cfg *config.CloudConfig, name string, serviceConfig *composeConfig.ServiceConfigV1) (project.Service, error) {
	if cfg == nil {
		cfg = config.LoadConfig()
	}

	p, err := CreateServiceSet("once", cfg, map[string]*composeConfig.ServiceConfigV1{
		name: serviceConfig,
	})
	if err != nil {
		return nil, err
	}

	return p.CreateService(name)
}

func CreateServiceSet(name string, cfg *config.CloudConfig, configs map[string]*composeConfig.ServiceConfigV1) (*project.Project, error) {
	p, err := newProject(name, cfg, nil, nil)
	if err != nil {
		return nil, err
	}

	addServices(p, map[interface{}]interface{}{}, configs)

	return p, nil
}

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

func preprocessServiceMap(serviceMap composeConfig.RawServiceMap) (composeConfig.RawServiceMap, error) {
	newServiceMap := make(composeConfig.RawServiceMap)

	for k, v := range serviceMap {
		newServiceMap[k] = make(composeConfig.RawService)

		for k2, v2 := range v {
			if k2 == "environment" || k2 == "labels" {
				newServiceMap[k][k2] = preprocess(v2, true)
			} else {
				newServiceMap[k][k2] = preprocess(v2, false)
			}

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

func addServices(p *project.Project, enabled map[interface{}]interface{}, configs map[string]*composeConfig.ServiceConfigV1) map[interface{}]interface{} {
	serviceConfigsV2, _ := composeConfig.ConvertServices(configs)

	// Note: we ignore errors while loading services
	unchanged := true
	for name, serviceConfig := range serviceConfigsV2 {
		hash := composeConfig.GetServiceHash(name, serviceConfig)

		if enabled[name] == hash {
			continue
		}

		if err := p.AddConfig(name, serviceConfig); err != nil {
			log.Infof("Failed loading service %s", name)
			continue
		}

		if unchanged {
			enabled = util.MapCopy(enabled)
			unchanged = false
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

	for _, service := range services {
		bytes, err := network.LoadServiceResource(service, true, cfg)
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
	for _, serviceName := range p.ServiceConfigs.Keys() {
		serviceConfig, _ := p.ServiceConfigs.Get(serviceName)
		p.ServiceConfigs.Add(serviceName, &composeConfig.ServiceConfig{
			Image:  serviceConfig.Image,
			Labels: serviceConfig.Labels,
		})
	}

	return p.Pull(context.Background())
}
