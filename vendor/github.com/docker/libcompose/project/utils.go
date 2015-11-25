package project

import (
	"strings"

	"github.com/docker/docker/runconfig"
)

func DefaultDependentServices(p *Project, s Service) []ServiceRelationship {
	config := s.Config()
	if config == nil {
		return []ServiceRelationship{}
	}

	result := []ServiceRelationship{}
	for _, link := range config.Links.Slice() {
		result = append(result, NewServiceRelationship(link, REL_TYPE_LINK))
	}

	for _, volumesFrom := range config.VolumesFrom {
		result = append(result, NewServiceRelationship(volumesFrom, REL_TYPE_VOLUMES_FROM))
	}

	result = appendNs(p, result, s.Config().Net, REL_TYPE_NET_NAMESPACE)
	result = appendNs(p, result, s.Config().Ipc, REL_TYPE_IPC_NAMESPACE)

	return result
}

func appendNs(p *Project, rels []ServiceRelationship, conf string, relType ServiceRelationshipType) []ServiceRelationship {
	service := GetContainerFromIpcLikeConfig(p, conf)
	if service != "" {
		rels = append(rels, NewServiceRelationship(service, relType))
	}
	return rels
}

func NameAlias(name string) (string, string) {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	} else {
		return parts[0], parts[0]
	}
}

func GetContainerFromIpcLikeConfig(p *Project, conf string) string {
	ipc := runconfig.IpcMode(conf)
	if !ipc.IsContainer() {
		return ""
	}

	name := ipc.Container()
	if name == "" {
		return ""
	}

	if _, ok := p.Configs[name]; ok {
		return name
	} else {
		return ""
	}
}
