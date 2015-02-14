package docker

import (
	"time"

	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
)

const (
	MAX_WAIT = 30000
	INTERVAL = 100
)

func NewClient(cfg *config.Config) (*dockerClient.Client, error) {
	endpoint := "unix:///var/run/docker.sock"
	if cfg != nil {
		endpoint = cfg.DockerEndpoint
	}
	client, err := dockerClient.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	for i := 0; i < (MAX_WAIT / INTERVAL); i++ {
		_, err = client.Info()
		if err == nil {
			break
		}

		time.Sleep(INTERVAL * time.Millisecond)
	}

	if err != nil {
		return nil, err
	}

	return client, nil
}
