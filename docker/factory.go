package docker

import (
	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/project"
)

type ContainerFactory struct {
	cfg *config.Config
}

type containerBasedService struct {
	name          string
	project       *project.Project
	container     *Container
	serviceConfig *project.ServiceConfig
	cfg           *config.Config
}

func NewContainerFactory(cfg *config.Config) *ContainerFactory {
	return &ContainerFactory{
		cfg: cfg,
	}
}

func (c *containerBasedService) Up() error {
	container := c.container
	containerCfg := c.container.ContainerCfg

	create := containerCfg.CreateOnly

	if util.Contains(c.cfg.Disable, c.name) {
		create = true
	}

	var event project.Event

	if create {
		container.Create()
		event = project.CONTAINER_CREATED
	} else {
		container.StartAndWait()
		event = project.CONTAINER_STARTED
	}

	if container.Err != nil {
		log.Errorf("Failed to run %v: %v", containerCfg.Id, container.Err)
	}

	if container.Err == nil && containerCfg.ReloadConfig {
		return project.ErrRestart
	}

	if container.Container != nil {
		c.project.Notify(event, c, map[string]string{
			project.CONTAINER_ID: container.Container.ID,
		})
	}

	return container.Err
}

func (c *containerBasedService) Config() *project.ServiceConfig {
	return c.serviceConfig
}

func (c *containerBasedService) Name() string {
	return c.name
}

func (c *ContainerFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	container := NewContainerFromService(config.DOCKER_SYSTEM_HOST, name, serviceConfig)

	if container.Err != nil {
		return nil, container.Err
	}

	return &containerBasedService{
		name:          name,
		project:       project,
		container:     container,
		serviceConfig: serviceConfig,
		cfg:           c.cfg,
	}, nil
}
