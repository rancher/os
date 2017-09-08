package glue

import (
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
)

type DockerPluginState struct {
	ContainerID string
	HostConfig  container.HostConfig
	Config      container.Config
	Pid         int
}

func LookupPluginState(container types.ContainerJSON) (*DockerPluginState, error) {
	result := &DockerPluginState{}

	result.ContainerID = container.ID
	result.HostConfig = *container.HostConfig
	result.Config = *container.Config
	if container.State != nil {
		result.Pid = container.State.Pid
	}

	return result, nil
}
