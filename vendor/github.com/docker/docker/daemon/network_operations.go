package daemon

import (
	"fmt"
	"os"

	"github.com/docker/docker/container"
	derr "github.com/docker/docker/errors"
	networktypes "github.com/docker/engine-api/types/network"
)

func (daemon *Daemon) updateContainerNetworkSettings(container *container.Container, endpointsConfig map[string]*networktypes.EndpointSettings) error {
	return nil
}

func (daemon *Daemon) updateNetwork(container *container.Container) error {
	return nil
}

func (daemon *Daemon) allocateNetwork(container *container.Container) error {
	return nil
}

func (daemon *Daemon) releaseNetwork(container *container.Container) {
	if container.HostConfig.NetworkMode.IsContainer() || container.Config.NetworkDisabled {
		return
	}
}

func (daemon *Daemon) getNetworkedContainer(containerID, connectedContainerID string) (*container.Container, error) {
	nc, err := daemon.GetContainer(connectedContainerID)
	if err != nil {
		return nil, err
	}
	if containerID == nc.ID {
		return nil, fmt.Errorf("cannot join own network")
	}
	if !nc.IsRunning() {
		err := fmt.Errorf("cannot join network of a non running container: %s", connectedContainerID)
		return nil, derr.NewRequestConflictError(err)
	}
	if nc.IsRestarting() {
		return nil, errContainerIsRestarting(connectedContainerID)
	}
	return nc, nil
}

func (daemon *Daemon) initializeNetworking(container *container.Container) error {
	var err error

	if container.HostConfig.NetworkMode.IsContainer() {
		// we need to get the hosts files from the container to join
		nc, err := daemon.getNetworkedContainer(container.ID, container.HostConfig.NetworkMode.ConnectedContainer())
		if err != nil {
			return err
		}
		container.HostnamePath = nc.HostnamePath
		container.HostsPath = nc.HostsPath
		container.ResolvConfPath = nc.ResolvConfPath
		container.Config.Hostname = nc.Config.Hostname
		container.Config.Domainname = nc.Config.Domainname
		return nil
	}

	if container.HostConfig.NetworkMode.IsHost() {
		container.Config.Hostname, err = os.Hostname()
		if err != nil {
			return err
		}
	}

	if err := daemon.allocateNetwork(container); err != nil {
		return err
	}

	return container.BuildHostnameFile()
}
