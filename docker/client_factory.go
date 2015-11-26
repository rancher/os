package docker

import (
	"fmt"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

type ClientFactory struct {
	userClient   *dockerclient.Client
	systemClient *dockerclient.Client
	userOnce     sync.Once
	systemOnce   sync.Once
}

func NewClientFactory(opts docker.ClientOpts) (docker.ClientFactory, error) {
	userOpts := opts
	systemOpts := opts

	userOpts.Host = config.DOCKER_HOST
	systemOpts.Host = config.DOCKER_SYSTEM_HOST

	userClient, err := docker.CreateClient(userOpts)
	if err != nil {
		return nil, err
	}

	systemClient, err := docker.CreateClient(systemOpts)
	if err != nil {
		return nil, err
	}

	return &ClientFactory{
		userClient:   userClient,
		systemClient: systemClient,
	}, nil
}

func (c *ClientFactory) Create(service project.Service) *dockerclient.Client {
	if IsSystemContainer(service.Config()) {
		waitFor(&c.systemOnce, c.systemClient, config.DOCKER_SYSTEM_HOST)
		return c.systemClient
	}

	waitFor(&c.userOnce, c.userClient, config.DOCKER_HOST)
	return c.userClient
}

func waitFor(once *sync.Once, client *dockerclient.Client, endpoint string) {
	once.Do(func() {
		err := ClientOK(endpoint, func() bool {
			_, err := client.Info()
			return err == nil
		})
		if err != nil {
			panic(err.Error())
		}
	})
}

func ClientOK(endpoint string, test func() bool) error {
	backoff := util.Backoff{}
	defer backoff.Close()

	var err error
	retry := false
	for ok := range backoff.Start() {
		if !ok {
			err = fmt.Errorf("Timeout waiting for Docker at %s", endpoint)
			break
		}
		if test() {
			break
		}
		retry = true
		log.Infof("Waiting for Docker at %s", endpoint)
	}

	if err != nil {
		return err
	}

	if retry {
		log.Infof("Connected to Docker at %s", endpoint)
	}

	return nil
}
