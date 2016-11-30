package docker

import (
	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config"
)

func IsSystemContainer(serviceConfig *composeConfig.ServiceConfig) bool {
	return serviceConfig.Labels[config.ScopeLabel] == config.System
}
