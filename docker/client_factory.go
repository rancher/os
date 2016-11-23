package docker

import (
	"fmt"
	"sync"

	"golang.org/x/net/context"

	dockerclient "github.com/docker/engine-api/client"
	composeClient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
)

type ClientFactory struct {
	userClient   dockerclient.APIClient
	systemClient dockerclient.APIClient
	userOnce     sync.Once
	systemOnce   sync.Once
}

func NewClientFactory(opts composeClient.Options) (project.ClientFactory, error) {
	userOpts := opts
	systemOpts := opts

	userOpts.Host = config.DockerHost
	systemOpts.Host = config.SystemDockerHost

	userClient, err := composeClient.Create(userOpts)
	if err != nil {
		return nil, err
	}

	systemClient, err := composeClient.Create(systemOpts)
	if err != nil {
		return nil, err
	}

	return &ClientFactory{
		userClient:   userClient,
		systemClient: systemClient,
	}, nil
}

func (c *ClientFactory) Create(service project.Service) dockerclient.APIClient {
	if IsSystemContainer(service.Config()) {
		waitFor(&c.systemOnce, c.systemClient, config.SystemDockerHost)
		return c.systemClient
	}

	waitFor(&c.userOnce, c.userClient, config.DockerHost)
	return c.userClient
}

func waitFor(once *sync.Once, client dockerclient.APIClient, endpoint string) {
	once.Do(func() {
		err := ClientOK(endpoint, func() bool {
			_, err := client.Info(context.Background())
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
