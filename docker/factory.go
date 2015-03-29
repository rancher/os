package docker

import (
	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/config"
	"github.com/rancherio/rancher-compose/project"
)

type ContainerFactory struct {
}

type containerBasedService struct {
	name          string
	project       *project.Project
	container     *Container
	serviceConfig *project.ServiceConfig
	cfg           *config.Config
}

func (c *containerBasedService) Up() error {
	container := c.container
	containerCfg := c.container.ContainerCfg

	if containerCfg.CreateOnly {
		container.Create()
		c.project.Notify(project.CONTAINER_CREATED, c, map[string]string{
			project.CONTAINER_ID: container.Container.ID,
		})
	} else {
		container.StartAndWait()
		c.project.Notify(project.CONTAINER_STARTED, c, map[string]string{
			project.CONTAINER_ID: container.Container.ID,
		})
	}

	if container.Err != nil {
		log.Errorf("Failed to run %v: %v", containerCfg.Id, container.Err)
	}

	if container.Err == nil && containerCfg.ReloadConfig {
		return project.ErrRestart
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
	}, nil
}
