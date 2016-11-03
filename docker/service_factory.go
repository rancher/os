package docker

import (
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/util"
)

type ServiceFactory struct {
	Context *docker.Context
	Deps    map[string][]string
}

func (s *ServiceFactory) Create(project *project.Project, name string, serviceConfig *composeConfig.ServiceConfig) (project.Service, error) {
	if after := serviceConfig.Labels["io.rancher.os.after"]; after != "" {
		for _, dep := range util.TrimSplit(after, ",") {
			if dep == "cloud-init" {
				dep = "cloud-init-execute"
			}
			s.Deps[name] = append(s.Deps[name], dep)
		}
	}
	if before := serviceConfig.Labels["io.rancher.os.before"]; before != "" {
		for _, dep := range util.TrimSplit(before, ",") {
			if dep == "cloud-init" {
				dep = "cloud-init-execute"
			}
			s.Deps[dep] = append(s.Deps[dep], name)
		}
	}

	return NewService(s, name, serviceConfig, s.Context, project), nil
}
