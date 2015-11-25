package docker

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/utils"
)

type Service struct {
	name          string
	serviceConfig *project.ServiceConfig
	context       *Context
}

func NewService(name string, serviceConfig *project.ServiceConfig, context *Context) *Service {
	return &Service{
		name:          name,
		serviceConfig: serviceConfig,
		context:       context,
	}
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Config() *project.ServiceConfig {
	return s.serviceConfig
}

func (s *Service) DependentServices() []project.ServiceRelationship {
	return project.DefaultDependentServices(s.context.Project, s)
}

func (s *Service) Create() error {
	imageName, err := s.build()
	if err != nil {
		return err
	}

	_, err = s.createOne(imageName)
	return err
}

func (s *Service) collectContainers() ([]*Container, error) {
	client := s.context.ClientFactory.Create(s)
	containers, err := GetContainersByFilter(client, SERVICE.Eq(s.name), PROJECT.Eq(s.context.Project.Name))
	if err != nil {
		return nil, err
	}

	result := []*Container{}

	if len(containers) == 0 {
		return result, nil
	}

	for _, container := range containers {
		name := container.Labels[NAME.Str()]
		result = append(result, NewContainer(client, name, s))
	}

	return result, nil
}

func (s *Service) createOne(imageName string) (*Container, error) {
	containers, err := s.constructContainers(imageName, 1)
	if err != nil {
		return nil, err
	}

	return containers[0], err
}

func (s *Service) Build() error {
	_, err := s.build()
	return err
}

func (s *Service) build() (string, error) {
	if s.context.Builder == nil {
		return s.Config().Image, nil
	}

	return s.context.Builder.Build(s.context.Project, s)
}

func (s *Service) constructContainers(imageName string, count int) ([]*Container, error) {
	result, err := s.collectContainers()
	if err != nil {
		return nil, err
	}

	client := s.context.ClientFactory.Create(s)

	namer := NewNamer(client, s.context.Project.Name, s.name)
	defer namer.Close()

	for i := len(result); i < count; i++ {
		containerName := namer.Next()

		c := NewContainer(client, containerName, s)

		dockerContainer, err := c.Create(imageName)
		if err != nil {
			return nil, err
		}

		logrus.Debugf("Created container %s: %v", dockerContainer.Id, dockerContainer.Names)

		result = append(result, NewContainer(client, containerName, s))
	}

	return result, nil
}

func (s *Service) Up() error {
	imageName, err := s.build()
	if err != nil {
		return err
	}

	return s.up(imageName, true)
}

func (s *Service) Info() (project.InfoSet, error) {
	result := project.InfoSet{}
	containers, err := s.collectContainers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		if info, err := c.Info(); err != nil {
			return nil, err
		} else {
			result = append(result, info)
		}
	}

	return result, nil
}

func (s *Service) Start() error {
	return s.up("", false)
}

func (s *Service) up(imageName string, create bool) error {
	containers, err := s.collectContainers()
	if err != nil {
		return err
	}

	logrus.Debugf("Found %d existing containers for service %s", len(containers), s.name)

	if len(containers) == 0 && create {
		c, err := s.createOne(imageName)
		if err != nil {
			return err
		}
		containers = []*Container{c}
	}

	return s.eachContainer(func(c *Container) error {
		if s.context.Rebuild && create {
			if err := s.rebuildIfNeeded(imageName, c); err != nil {
				return err
			}
		}

		return c.Up(imageName)
	})
}

func (s *Service) rebuildIfNeeded(imageName string, c *Container) error {
	outOfSync, err := c.OutOfSync(imageName)
	if err != nil {
		return err
	}

	containerInfo, err := c.findInfo()
	if containerInfo == nil || err != nil {
		return err
	}
	name := containerInfo.Name[1:]

	origRebuildLabel := containerInfo.Config.Labels[REBUILD.Str()]
	newRebuildLabel := s.Config().Labels.MapParts()[REBUILD.Str()]
	rebuildLabelChanged := newRebuildLabel != origRebuildLabel
	logrus.WithFields(logrus.Fields{
		"origRebuildLabel":    origRebuildLabel,
		"newRebuildLabel":     newRebuildLabel,
		"rebuildLabelChanged": rebuildLabelChanged,
		"outOfSync":           outOfSync}).Debug("Rebuild values")

	if origRebuildLabel == "always" || rebuildLabelChanged || origRebuildLabel != "false" && outOfSync {
		logrus.Infof("Rebuilding %s", name)
		if _, err := c.Rebuild(imageName); err != nil {
			return err
		}
	} else if outOfSync {
		logrus.Warnf("%s needs rebuilding", name)
	}

	return nil
}

func (s *Service) eachContainer(action func(*Container) error) error {
	containers, err := s.collectContainers()
	if err != nil {
		return err
	}

	tasks := utils.InParallel{}
	for _, container := range containers {
		task := func(container *Container) func() error {
			return func() error {
				return action(container)
			}
		}(container)

		tasks.Add(task)
	}

	return tasks.Wait()
}

func (s *Service) Down() error {
	return s.eachContainer(func(c *Container) error {
		return c.Down()
	})
}

func (s *Service) Restart() error {
	return s.eachContainer(func(c *Container) error {
		return c.Restart()
	})
}

func (s *Service) Kill() error {
	return s.eachContainer(func(c *Container) error {
		return c.Kill()
	})
}

func (s *Service) Delete() error {
	return s.eachContainer(func(c *Container) error {
		return c.Delete()
	})
}

func (s *Service) Log() error {
	return s.eachContainer(func(c *Container) error {
		return c.Log()
	})
}

func (s *Service) Scale(scale int) error {
	foundCount := 0
	err := s.eachContainer(func(c *Container) error {
		foundCount++
		if foundCount > scale {
			err := c.Down()
			if err != nil {
				return err
			}

			return c.Delete()
		}
		return nil
	})

	if err != nil {
		return err
	}

	if foundCount != scale {
		imageName, err := s.build()
		if err != nil {
			return err
		}

		if _, err = s.constructContainers(imageName, scale); err != nil {
			return err
		}
	}

	return s.up("", false)
}

func (s *Service) Pull() error {
	if s.Config().Image == "" {
		return nil
	}

	return PullImage(s.context.ClientFactory.Create(s), s, s.Config().Image)
}

func (s *Service) Containers() ([]project.Container, error) {
	result := []project.Container{}
	containers, err := s.collectContainers()
	if err != nil {
		return nil, err
	}

	for _, c := range containers {
		result = append(result, c)
	}

	return result, nil
}
