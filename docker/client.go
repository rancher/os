package docker

import (
	"time"

	log "github.com/Sirupsen/logrus"

	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
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

	retry := false
	for i := 0; i < (MAX_WAIT / INTERVAL); i++ {
		_, err = client.Info()
		if err == nil {
			break
		}

		retry = true

		log.Infof("Waiting for Docker at %s", endpoint)
		time.Sleep(INTERVAL * time.Millisecond)
	}

	if err != nil {
		return nil, err
	}

	if retry {
		log.Infof("Connected to Docker at %s", endpoint)
	}
	return client, nil
}
