package daemon

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/runconfig"
	"github.com/opencontainers/runc/libcontainer/label"
)

func (daemon *Daemon) ContainerCreate(name string, config *runconfig.Config, hostConfig *runconfig.HostConfig) (string, []string, error) {
	if config == nil {
		return "", nil, fmt.Errorf("Config cannot be empty in order to create a container")
	}

	daemon.adaptContainerSettings(hostConfig)
	warnings, err := daemon.verifyContainerSettings(hostConfig, config)
	if err != nil {
		return "", warnings, err
	}

	container, buildWarnings, err := daemon.Create(config, hostConfig, name)
	if err != nil {
		if daemon.Graph().IsNotExist(err, config.Image) {
			_, tag := parsers.ParseRepositoryTag(config.Image)
			if tag == "" {
				tag = graph.DefaultTag
			}
			return "", warnings, fmt.Errorf("No such image: %s (tag: %s)", config.Image, tag)
		}
		return "", warnings, err
	}

	warnings = append(warnings, buildWarnings...)

	return container.ID, warnings, nil
}

// Create creates a new container from the given configuration with a given name.
func (daemon *Daemon) Create(config *runconfig.Config, hostConfig *runconfig.HostConfig, name string) (*Container, []string, error) {
	var (
		container *Container
		warnings  []string
		img       *image.Image
		imgID     string
		err       error
	)

	if config.Image != "" {
		img, err = daemon.repositories.LookupImage(config.Image)
		if err != nil {
			return nil, nil, err
		}
		if err = daemon.graph.CheckDepth(img); err != nil {
			return nil, nil, err
		}
		imgID = img.ID
	}

	if err := daemon.mergeAndVerifyConfig(config, img); err != nil {
		return nil, nil, err
	}
	if !config.NetworkDisabled && daemon.SystemConfig().IPv4ForwardingDisabled {
		warnings = append(warnings, "IPv4 forwarding is disabled.")
	}
	if hostConfig == nil {
		hostConfig = &runconfig.HostConfig{}
	}
	if hostConfig.SecurityOpt == nil {
		hostConfig.SecurityOpt, err = daemon.GenerateSecurityOpt(hostConfig.IpcMode, hostConfig.PidMode)
		if err != nil {
			return nil, nil, err
		}
	}
	if container, err = daemon.newContainer(name, config, imgID); err != nil {
		return nil, nil, err
	}
	if err := daemon.Register(container); err != nil {
		return nil, nil, err
	}
	if err := daemon.createRootfs(container); err != nil {
		return nil, nil, err
	}
	if err := daemon.setHostConfig(container, hostConfig); err != nil {
		return nil, nil, err
	}
	if err := container.Mount(); err != nil {
		return nil, nil, err
	}
	defer container.Unmount()

	if err := createContainerPlatformSpecificSettings(container, config); err != nil {
		return nil, nil, err
	}

	if err := container.ToDisk(); err != nil {
		logrus.Errorf("Error saving new container to disk: %v", err)
		return nil, nil, err
	}
	container.LogEvent("create")
	return container, warnings, nil
}

func (daemon *Daemon) GenerateSecurityOpt(ipcMode runconfig.IpcMode, pidMode runconfig.PidMode) ([]string, error) {
	if ipcMode.IsHost() || pidMode.IsHost() {
		return label.DisableSecOpt(), nil
	}
	if ipcContainer := ipcMode.Container(); ipcContainer != "" {
		c, err := daemon.Get(ipcContainer)
		if err != nil {
			return nil, err
		}

		return label.DupSecOpt(c.ProcessLabel), nil
	}
	return nil, nil
}
