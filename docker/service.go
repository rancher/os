package docker

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/rancher/os/config"
)

type Service struct {
	*docker.Service
	deps    map[string][]string
	context *docker.Context
	project *project.Project
}

func NewService(factory *ServiceFactory, name string, serviceConfig *project.ServiceConfig, context *docker.Context, project *project.Project) *Service {
	return &Service{
		Service: docker.NewService(name, serviceConfig, context),
		deps:    factory.Deps,
		context: context,
		project: project,
	}
}

func (s *Service) DependentServices() []project.ServiceRelationship {
	rels := s.Service.DependentServices()
	for _, dep := range s.deps[s.Name()] {
		rels = appendLink(rels, dep, true, s.project)
	}

	if s.requiresSyslog() {
		rels = appendLink(rels, "syslog", false, s.project)
	}

	if s.requiresUserDocker() {
		// Linking to cloud-init is a hack really.  The problem is we need to link to something
		// that will trigger a reload
		rels = appendLink(rels, "cloud-init", false, s.project)
	} else if s.missingImage() {
		rels = appendLink(rels, "network", false, s.project)
	}
	return rels
}

func (s *Service) missingImage() bool {
	image := s.Config().Image
	if image == "" {
		return false
	}
	client := s.context.ClientFactory.Create(s)
	i, err := client.InspectImage(s.Config().Image)
	return err != nil || i == nil
}

func (s *Service) requiresSyslog() bool {
	return s.Config().LogDriver == "syslog"
}

func (s *Service) requiresUserDocker() bool {
	return s.Config().Labels.MapParts()[config.SCOPE] != config.SYSTEM
}

func appendLink(deps []project.ServiceRelationship, name string, optional bool, p *project.Project) []project.ServiceRelationship {
	if _, ok := p.Configs[name]; !ok {
		return deps
	}
	rel := project.NewServiceRelationship(name, project.RelTypeLink)
	rel.Optional = optional
	return append(deps, rel)
}

func (s *Service) shouldRebuild() (bool, error) {
	containers, err := s.Containers()
	if err != nil {
		return false, err
	}
	for _, c := range containers {
		outOfSync, err := c.(*docker.Container).OutOfSync(s.Service.Config().Image)
		if err != nil {
			return false, err
		}

		_, containerInfo, err := s.getContainer()
		if containerInfo == nil || err != nil {
			return false, err
		}
		name := containerInfo.Name[1:]

		origRebuildLabel := containerInfo.Config.Labels[config.REBUILD]
		newRebuildLabel := s.Config().Labels.MapParts()[config.REBUILD]
		rebuildLabelChanged := newRebuildLabel != origRebuildLabel
		logrus.WithFields(logrus.Fields{
			"origRebuildLabel":    origRebuildLabel,
			"newRebuildLabel":     newRebuildLabel,
			"rebuildLabelChanged": rebuildLabelChanged,
			"outOfSync":           outOfSync}).Debug("Rebuild values")

		if origRebuildLabel == "always" || rebuildLabelChanged || origRebuildLabel != "false" && outOfSync {
			logrus.Infof("Rebuilding %s", name)
			return true, err
		} else if outOfSync {
			logrus.Warnf("%s needs rebuilding", name)
		}
	}
	return false, nil
}

func (s *Service) Up() error {
	labels := s.Config().Labels.MapParts()

	if err := s.Service.Create(); err != nil {
		return err
	}
	shouldRebuild, err := s.shouldRebuild()
	if err != nil {
		return err
	}
	if shouldRebuild {
		cs, err := s.Service.Containers()
		if err != nil {
			return err
		}
		for _, c := range cs {
			if _, err := c.(*docker.Container).Recreate(s.Config().Image); err != nil {
				return err
			}
		}
		s.rename()
	}
	if labels[config.CREATE_ONLY] == "true" {
		return s.checkReload(labels)
	}
	if err := s.Service.Up(); err != nil {
		return err
	}
	if labels[config.DETACH] == "false" {
		if err := s.wait(); err != nil {
			return err
		}
	}

	return s.checkReload(labels)
}

func (s *Service) checkReload(labels map[string]string) error {
	if labels[config.RELOAD_CONFIG] == "true" {
		return project.ErrRestart
	}
	return nil
}

func (s *Service) Create() error {
	return s.Service.Create()
}

func (s *Service) getContainer() (*dockerclient.Client, *dockerclient.Container, error) {
	containers, err := s.Service.Containers()
	if err != nil {
		return nil, nil, err
	}

	if len(containers) == 0 {
		return nil, nil, nil
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, nil, err
	}

	client := s.context.ClientFactory.Create(s)
	info, err := client.InspectContainer(id)
	return client, info, err
}

func (s *Service) wait() error {
	client, info, err := s.getContainer()
	if err != nil || info == nil {
		return err
	}

	if _, err := client.WaitContainer(info.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) rename() error {
	client, info, err := s.getContainer()
	if err != nil || info == nil {
		return err
	}

	if len(info.Name) > 0 && info.Name[1:] != s.Name() {
		logrus.Debugf("Renaming container %s => %s", info.Name[1:], s.Name())
		return client.RenameContainer(dockerclient.RenameContainerOptions{ID: info.ID, Name: s.Name()})
	} else {
		return nil
	}
}
