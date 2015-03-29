package control

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

//func parseContainers(cfg *config.Config) map[string]*docker.Container {
//	result := map[string]*docker.Container{}
//
//	for _, containerConfig := range cfg.SystemContainers {
//		container := docker.NewContainer(config.DOCKER_SYSTEM_HOST, &containerConfig)
//		if containerConfig.Id != "" {
//			result[containerConfig.Id] = container
//		}
//	}
//
//	return result
//}

func reload(c *cli.Context) {
	_, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	containers := map[string]*docker.Container{} //parseContainers(cfg)
	toStart := make([]*docker.Container, 0, len(c.Args()))

	for _, id := range c.Args() {
		if container, ok := containers[id]; ok {
			toStart = append(toStart, container.Stage())
		}
	}

	var firstErr error
	for _, c := range toStart {
		err := c.Start().Err
		if err != nil {
			log.Errorf("Failed to start %s : %v", c.ContainerCfg.Id, err)
			if firstErr != nil {
				firstErr = err
			}
		}
	}

	if firstErr != nil {
		log.Fatal(firstErr)
	}
}
