package docker

import (
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/util"
)

type ServiceFactory struct {
	Context *docker.Context
	Deps    map[string][]string
}

func (s *ServiceFactory) Create(project *project.Project, name string, serviceConfig *project.ServiceConfig) (project.Service, error) {
	if after := serviceConfig.Labels.MapParts()["io.rancher.os.after"]; after != "" {
		for _, dep := range util.TrimSplit(after, ",") {
			s.Deps[name] = append(s.Deps[name], dep)
		}
	}
	if before := serviceConfig.Labels.MapParts()["io.rancher.os.before"]; before != "" {
		for _, dep := range util.TrimSplit(before, ",") {
			s.Deps[dep] = append(s.Deps[dep], name)
		}
	}

	return NewService(s, name, serviceConfig, s.Context, project), nil
}
