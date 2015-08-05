package docker

import (
	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

type ContainerFactory struct {
	cfg *config.CloudConfig
}

type containerBasedService struct {
	project.EmptyService
	name          string
	project       *project.Project
	container     *Container
	serviceConfig *project.ServiceConfig
	cfg           *config.CloudConfig
}

func NewContainerFactory(cfg *config.CloudConfig) *ContainerFactory {
	return &ContainerFactory{
		cfg: cfg,
	}
}

func (c *containerBasedService) Up() error {
	container := c.container
	containerCfg := c.container.ContainerCfg

	fakeCreate := false
	create := containerCfg.CreateOnly

	if util.Contains(c.cfg.Rancher.Disable, c.name) {
		fakeCreate = true
	}

	var event project.Event

	c.project.Notify(project.CONTAINER_STARTING, c.name, map[string]string{})

	if fakeCreate {
		event = project.CONTAINER_CREATED
	} else if create {
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
		c.project.Notify(event, c.name, map[string]string{
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

func isSystemService(serviceConfig *project.ServiceConfig) bool {
	return serviceConfig.Labels.MapParts()[config.SCOPE] == config.SYSTEM
}

func (c *ContainerFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	host := config.DOCKER_HOST
	if isSystemService(serviceConfig) {
		host = config.DOCKER_SYSTEM_HOST
	}

	container := NewContainerFromService(host, name, serviceConfig)

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
