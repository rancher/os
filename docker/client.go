package docker

import (
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancher/os/config"
)

const (
	MAX_WAIT = 30000
	INTERVAL = 100
)

func NewSystemClient() (*dockerClient.Client, error) {
	return NewClient(config.DOCKER_SYSTEM_HOST)
}

func NewDefaultClient() (*dockerClient.Client, error) {
	return NewClient(config.DOCKER_HOST)
}

func NewClient(endpoint string) (*dockerClient.Client, error) {
	client, err := dockerClient.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	err = ClientOK(endpoint, func() bool {
		_, err := client.Info()
		return err == nil
	})

	return client, err
}
