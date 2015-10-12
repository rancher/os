package docker

import (
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/config"
)

func IsSystemContainer(serviceConfig *project.ServiceConfig) bool {
	return serviceConfig.Labels.MapParts()[config.SCOPE] == config.SYSTEM

}
