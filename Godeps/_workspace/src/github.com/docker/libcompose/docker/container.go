package docker

import (
	"bufio"
	"fmt"
	"math"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/graph/tags"
	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/utils"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/project"
	"github.com/samalba/dockerclient"
	"os"
)

type Container struct {
	project.EmptyService

	name    string
	service *Service
	client  dockerclient.Client
}

func NewContainer(client dockerclient.Client, name string, service *Service) *Container {
	return &Container{
		client:  client,
		name:    name,
		service: service,
	}
}

func (c *Container) findExisting() (*dockerclient.Container, error) {
	return GetContainerByName(c.client, c.name)
}

func (c *Container) findInfo() (*dockerclient.ContainerInfo, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	return c.client.InspectContainer(container.Id)
}

func (c *Container) Info() (project.Info, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	result := project.Info{}

	result = append(result, project.InfoPart{Key: "Name", Value: name(container.Names)})
	result = append(result, project.InfoPart{Key: "Command", Value: container.Command})
	result = append(result, project.InfoPart{Key: "State", Value: container.Status})
	result = append(result, project.InfoPart{Key: "Ports", Value: portString(container.Ports)})

	return result, nil
}

func portString(ports []dockerclient.Port) string {
	result := []string{}

	for _, port := range ports {
		if port.PublicPort > 0 {
			result = append(result, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
		} else {
			result = append(result, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}

	return strings.Join(result, ", ")
}

func name(names []string) string {
	max := math.MaxInt32
	var current string

	for _, v := range names {
		if len(v) < max {
			max = len(v)
			current = v
		}
	}

	return current[1:]
}

func (c *Container) Rebuild(imageName string) (*dockerclient.Container, error) {
	info, err := c.findInfo()
	if err != nil {
		return nil, err
	} else if info == nil {
		return nil, fmt.Errorf("Can not find container to rebuild for service: %s", c.service.Name())
	}

	hash := info.Config.Labels[HASH.Str()]
	if hash == "" {
		return nil, fmt.Errorf("Failed to find hash on old container: %s", info.Name)
	}

	name := info.Name[1:]
	new_name := fmt.Sprintf("%s_%s", name, info.Id[:12])
	deleted := false
	logrus.Debugf("Renaming %s => %s", name, new_name)
	if err := c.client.RenameContainer(name, new_name); err != nil {
		logrus.Errorf("Rename failed, deleting %s", name)
		if err := c.client.RemoveContainer(info.Id, true, false); err != nil {
			return nil, err
		}
		deleted = true
	}

	newContainer, err := c.createContainer(imageName, info.Id)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Created replacement container %s", newContainer.Id)

	if !deleted {
		if err := c.client.RemoveContainer(info.Id, true, false); err != nil {
			logrus.Errorf("Failed to remove old container %s", c.name)
			return nil, err
		}
		logrus.Debugf("Removed old container %s %s", c.name, info.Id)
	}

	return newContainer, nil
}

func (c *Container) Create(imageName string) (*dockerclient.Container, error) {
	container, err := c.findExisting()
	if err != nil {
		return nil, err
	}

	if container == nil {
		container, err = c.createContainer(imageName, "")
		if err != nil {
			return nil, err
		}
		c.service.context.Project.Notify(project.CONTAINER_CREATED, c.service.Name(), map[string]string{
			"name": c.Name(),
		})
	}

	return container, err
}

func (c *Container) Down() error {
	return c.withContainer(func(container *dockerclient.Container) error {
		return c.client.StopContainer(container.Id, c.service.context.Timeout)
	})
}

func (c *Container) Kill() error {
	return c.withContainer(func(container *dockerclient.Container) error {
		return c.client.KillContainer(container.Id, c.service.context.Signal)
	})
}

func (c *Container) Delete() error {
	container, err := c.findExisting()
	if err != nil || container == nil {
		return err
	}

	info, err := c.client.InspectContainer(container.Id)
	if err != nil {
		return err
	}

	if info.State.Running {
		err := c.client.StopContainer(container.Id, c.service.context.Timeout)
		if err != nil {
			return err
		}
	}

	return c.client.RemoveContainer(container.Id, true, false)
}

func (c *Container) Up(imageName string) error {
	var err error

	defer func() {
		if err == nil && c.service.context.Log {
			go c.Log()
		}
	}()

	container, err := c.Create(imageName)
	if err != nil {
		return err
	}

	info, err := c.client.InspectContainer(container.Id)
	if err != nil {
		return err
	}

	if !info.State.Running {
		logrus.Debugf("Starting container: %s", container.Id)
		if err := c.client.StartContainer(container.Id, nil); err != nil {
			return err
		}

		c.service.context.Project.Notify(project.CONTAINER_STARTED, c.service.Name(), map[string]string{
			"name": c.Name(),
		})
	}

	return nil
}

func (c *Container) OutOfSync(imageName string) (bool, error) {
	info, err := c.findInfo()
	if err != nil || info == nil {
		return false, err
	}

	if info.Config.Image != imageName {
		logrus.Debugf("Images for %s do not match %s!=%s", c.name, info.Config.Image, imageName)
		return true, nil
	}

	if info.Config.Labels[HASH.Str()] != c.getHash() {
		logrus.Debugf("Hashes for %s do not match %s!=%s", c.name, info.Config.Labels[HASH.Str()], c.getHash())
		return true, nil
	}

	image, err := c.client.InspectImage(info.Config.Image)
	if err != nil && (err.Error() == "Not found" || image == nil) {
		logrus.Debugf("Image %s do not exist, do not know if it's out of sync", info.Config.Image)
		return false, nil
	} else if err != nil {
		return false, err
	}

	logrus.Debugf("Checking existing image name vs id: %s == %s", image.Id, info.Image)
	return image.Id != info.Image, err
}

func (c *Container) getHash() string {
	return project.GetServiceHash(c.service.Name(), *c.service.Config())
}

func (c *Container) createContainer(imageName, oldContainer string) (*dockerclient.Container, error) {
	config, err := ConvertToApi(c.service.serviceConfig)
	if err != nil {
		return nil, err
	}

	config.Image = imageName

	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[NAME.Str()] = c.name
	config.Labels[SERVICE.Str()] = c.service.name
	config.Labels[PROJECT.Str()] = c.service.context.Project.Name
	config.Labels[HASH.Str()] = c.getHash()

	err = c.populateAdditionalHostConfig(&config.HostConfig)
	if err != nil {
		return nil, err
	}

	if oldContainer != "" {
		config.HostConfig.VolumesFrom = append(config.HostConfig.VolumesFrom, oldContainer)
	}

	logrus.Debugf("Creating container %s %#v", c.name, config)

	id, err := c.client.CreateContainer(config, c.name)
	if err != nil && err.Error() == "Not found" {
		logrus.Debugf("Not Found, pulling image %s", config.Image)
		if err = c.pull(config.Image); err != nil {
			return nil, err
		}
		if id, err = c.client.CreateContainer(config, c.name); err != nil {
			return nil, err
		}
	}

	if err != nil {
		logrus.Debugf("Failed to create container %s: %v", c.name, err)
		return nil, err
	}

	return GetContainerById(c.client, id)
}

func (c *Container) populateAdditionalHostConfig(hostConfig *dockerclient.HostConfig) error {
	links := map[string]string{}

	for _, link := range c.service.DependentServices() {
		if _, ok := c.service.context.Project.Configs[link.Target]; !ok {
			continue
		}

		service, err := c.service.context.Project.CreateService(link.Target)
		if err != nil {
			return err
		}

		containers, err := service.Containers()
		if err != nil {
			return err
		}

		if link.Type == project.REL_TYPE_LINK {
			c.addLinks(links, service, link, containers)
		} else if link.Type == project.REL_TYPE_IPC_NAMESPACE {
			hostConfig, err = c.addIpc(hostConfig, service, containers)
		} else if link.Type == project.REL_TYPE_NET_NAMESPACE {
			hostConfig, err = c.addNetNs(hostConfig, service, containers)
		}

		if err != nil {
			return err
		}
	}

	hostConfig.Links = []string{}
	for k, v := range links {
		hostConfig.Links = append(hostConfig.Links, strings.Join([]string{v, k}, ":"))
	}
	for _, v := range c.service.Config().ExternalLinks {
		hostConfig.Links = append(hostConfig.Links, v)
	}

	return nil
}

func (c *Container) addLinks(links map[string]string, service project.Service, rel project.ServiceRelationship, containers []project.Container) {
	for _, container := range containers {
		if _, ok := links[rel.Alias]; !ok {
			links[rel.Alias] = container.Name()
		}

		links[container.Name()] = container.Name()
	}
}

func (c *Container) addIpc(config *dockerclient.HostConfig, service project.Service, containers []project.Container) (*dockerclient.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for IPC %v", c.service.Config().Ipc)
	}

	id, err := containers[0].Id()
	if err != nil {
		return nil, err
	}

	config.IpcMode = "container:" + id
	return config, nil
}

func (c *Container) addNetNs(config *dockerclient.HostConfig, service project.Service, containers []project.Container) (*dockerclient.HostConfig, error) {
	if len(containers) == 0 {
		return nil, fmt.Errorf("Failed to find container for networks ns %v", c.service.Config().Net)
	}

	id, err := containers[0].Id()
	if err != nil {
		return nil, err
	}

	config.NetworkMode = "container:" + id
	return config, nil
}

func (c *Container) Id() (string, error) {
	container, err := c.findExisting()
	if container == nil {
		return "", err
	} else {
		return container.Id, err
	}
}

func (c *Container) Name() string {
	return c.name
}

func (c *Container) Pull() error {
	return c.pull(c.service.serviceConfig.Image)
}

func (c *Container) Restart() error {
	container, err := c.findExisting()
	if err != nil || container == nil {
		return err
	}

	return c.client.RestartContainer(container.Id, c.service.context.Timeout)
}

func (c *Container) Log() error {
	container, err := c.findExisting()
	if container == nil || err != nil {
		return err
	}

	info, err := c.client.InspectContainer(container.Id)
	if info == nil || err != nil {
		return err
	}

	l := c.service.context.LoggerFactory.Create(c.name)

	output, err := c.client.ContainerLogs(container.Id, &dockerclient.LogOptions{
		Follow: true,
		Stdout: true,
		Stderr: true,
		Tail:   10,
	})
	if err != nil {
		return err
	}

	if info.Config.Tty {
		scanner := bufio.NewScanner(output)
		for scanner.Scan() {
			l.Out([]byte(scanner.Text() + "\n"))
		}
		return scanner.Err()
	} else {
		_, err := stdcopy.StdCopy(&logger.LoggerWrapper{
			Logger: l,
		}, &logger.LoggerWrapper{
			Err:    true,
			Logger: l,
		}, output)
		return err
	}
}

func (c *Container) pull(image string) error {
	return PullImage(c.client, c.service, image)
}

func PullImage(client dockerclient.Client, service *Service, image string) error {
	taglessRemote, tag := parsers.ParseRepositoryTag(image)
	if tag == "" {
		image = utils.ImageReference(taglessRemote, tags.DEFAULTTAG)
	}

	repoInfo, err := registry.ParseRepositoryInfo(taglessRemote)
	if err != nil {
		return err
	}

	authConfig := cliconfig.AuthConfig{}
	if service.context.ConfigFile != nil && repoInfo != nil && repoInfo.Index != nil {
		authConfig = registry.ResolveAuthConfig(service.context.ConfigFile, repoInfo.Index)
	}

	err = client.PullImage(image, &dockerclient.AuthConfig{
		Username: authConfig.Username,
		Password: authConfig.Password,
		Email:    authConfig.Email,
	}, os.Stderr)

	if err != nil {
		logrus.Errorf("Failed to pull image %s: %v", image, err)
	}

	return err
}

func (c *Container) withContainer(action func(*dockerclient.Container) error) error {
	container, err := c.findExisting()
	if err != nil {
		return err
	}

	if container != nil {
		return action(container)
	}

	return nil
}

func (c *Container) Port(port string) (string, error) {
	info, err := c.findInfo()
	if err != nil {
		return "", err
	}

	if bindings, ok := info.NetworkSettings.Ports[port]; ok {
		result := []string{}
		for _, binding := range bindings {
			result = append(result, binding.HostIp+":"+binding.HostPort)
		}

		return strings.Join(result, "\n"), nil
	} else {
		return "", nil
	}
}
