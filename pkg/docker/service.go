package docker

import (
	"fmt"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"

	"github.com/docker/docker/layer"
	dockerclient "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"golang.org/x/net/context"
)

type Service struct {
	*docker.Service
	deps    map[string][]string
	context *docker.Context
	project *project.Project
}

func NewService(factory *ServiceFactory, name string, serviceConfig *composeConfig.ServiceConfig, context *docker.Context, project *project.Project) *Service {
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
		rels = appendLink(rels, "docker", false, s.project)
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
	_, _, err := client.ImageInspectWithRaw(context.Background(), image, false)
	return err != nil
}

func (s *Service) requiresSyslog() bool {
	return s.Config().Logging.Driver == "syslog"
}

func (s *Service) requiresUserDocker() bool {
	return s.Config().Labels[config.ScopeLabel] != config.System
}

func appendLink(deps []project.ServiceRelationship, name string, optional bool, p *project.Project) []project.ServiceRelationship {
	if _, ok := p.ServiceConfigs.Get(name); !ok {
		return deps
	}
	rel := project.NewServiceRelationship(name, project.RelTypeLink)
	rel.Optional = optional
	return append(deps, rel)
}

func (s *Service) shouldRebuild(ctx context.Context) (bool, error) {
	containers, err := s.Containers(ctx)
	if err != nil {
		return false, err
	}
	cfg := config.LoadConfig()
	for _, c := range containers {
		outOfSync, err := c.(*docker.Container).OutOfSync(ctx, s.Service.Config().Image)
		if err != nil {
			return false, err
		}

		_, containerInfo, err := s.getContainer(ctx)
		if err != nil {
			return false, err
		}
		name := containerInfo.Name[1:]

		origRebuildLabel := containerInfo.Config.Labels[config.RebuildLabel]
		newRebuildLabel := s.Config().Labels[config.RebuildLabel]
		rebuildLabelChanged := newRebuildLabel != origRebuildLabel
		log.WithFields(log.Fields{
			"origRebuildLabel":    origRebuildLabel,
			"newRebuildLabel":     newRebuildLabel,
			"rebuildLabelChanged": rebuildLabelChanged,
			"outOfSync":           outOfSync}).Debug("Rebuild values")

		if newRebuildLabel == "always" {
			return true, nil
		}
		if s.Name() == "console" && cfg.Rancher.ForceConsoleRebuild {
			if err := config.Set("rancher.force_console_rebuild", false); err != nil {
				return false, err
			}
			return true, nil
		}
		if outOfSync {
			if s.Name() == "console" {
				origConsoleLabel := containerInfo.Config.Labels[config.ConsoleLabel]
				newConsoleLabel := s.Config().Labels[config.ConsoleLabel]
				if newConsoleLabel != origConsoleLabel {
					return true, nil
				}
			} else if rebuildLabelChanged || origRebuildLabel != "false" {
				return true, nil
			} else {
				log.Warnf("%s needs rebuilding", name)
			}
		}
	}
	return false, nil
}

func (s *Service) Up(ctx context.Context, options options.Up) error {
	labels := s.Config().Labels

	if err := s.Service.Create(ctx, options.Create); err != nil {
		return err
	}

	shouldRebuild, err := s.shouldRebuild(ctx)
	if err != nil {
		return err
	}
	if shouldRebuild {
		log.Infof("Rebuilding %s", s.Name())
		cs, err := s.Service.Containers(ctx)
		if err != nil {
			return err
		}
		for _, c := range cs {
			if _, err := c.(*docker.Container).Recreate(ctx, s.Config().Image); err != nil {
				// sometimes we can get ErrMountNameConflict when booting on RPi
				// ignore this error so that ros can boot success, otherwise it will hang forever
				if strings.Contains(err.Error(), layer.ErrMountNameConflict.Error()) {
					log.Warn(err)
				} else {
					return err
				}
			}
		}
		if err = s.rename(ctx); err != nil {
			return err
		}
	}
	if labels[config.CreateOnlyLabel] == "true" {
		return s.checkReload(labels)
	}
	if err := s.Service.Up(ctx, options); err != nil {
		return err
	}
	if labels[config.DetachLabel] == "false" {
		if err := s.wait(ctx); err != nil {
			return err
		}
	}

	return s.checkReload(labels)
}

func (s *Service) checkReload(labels map[string]string) error {
	if labels[config.ReloadConfigLabel] == "true" {
		return project.ErrRestart
	}
	return nil
}

func (s *Service) Create(ctx context.Context, options options.Create) error {
	return s.Service.Create(ctx, options)
}

func (s *Service) getContainer(ctx context.Context) (dockerclient.APIClient, types.ContainerJSON, error) {
	containers, err := s.Service.Containers(ctx)

	if err != nil {
		return nil, types.ContainerJSON{}, err
	}

	if len(containers) == 0 {
		return nil, types.ContainerJSON{}, fmt.Errorf("No containers found for %s", s.Name())
	}

	id, err := containers[0].ID()
	if err != nil {
		return nil, types.ContainerJSON{}, err
	}

	client := s.context.ClientFactory.Create(s)
	info, err := client.ContainerInspect(context.Background(), id)
	return client, info, err
}

func (s *Service) wait(ctx context.Context) error {
	client, info, err := s.getContainer(ctx)
	if err != nil {
		return err
	}

	if _, err := client.ContainerWait(context.Background(), info.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) rename(ctx context.Context) error {
	client, info, err := s.getContainer(ctx)
	if err != nil {
		return err
	}

	if len(info.Name) > 0 && info.Name[1:] != s.Name() {
		log.Debugf("Renaming container %s => %s", info.Name[1:], s.Name())
		return client.ContainerRename(context.Background(), info.ID, s.Name())
	}
	return nil
}
