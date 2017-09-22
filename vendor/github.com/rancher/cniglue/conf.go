package glue

import (
	"encoding/json"
	"os"
	"path"

	"github.com/docker/engine-api/types/container"
	"github.com/opencontainers/specs/specs-go"
)

type DockerPluginState struct {
	ContainerID string
	State       specs.State
	Spec        specs.Spec
	HostConfig  container.HostConfig
	Config      container.Config
}

func ReadState() (*DockerPluginState, error) {
	pluginState := DockerPluginState{}
	config := struct {
		ID     string
		Config container.Config
	}{}

	if err := json.NewDecoder(os.Stdin).Decode(&pluginState.State); err != nil {
		return nil, err
	}

	if err := readJSONFile(os.Getenv("DOCKER_HOST_CONFIG"), &pluginState.HostConfig); err != nil {
		return nil, err
	}

	if err := readJSONFile(os.Getenv("DOCKER_CONFIG"), &config); err != nil {
		return nil, err
	}

	pluginState.Config = config.Config
	pluginState.ContainerID = config.ID

	return &pluginState, readJSONFile(path.Join(pluginState.State.BundlePath, "config.json"), &pluginState.Spec)
}
