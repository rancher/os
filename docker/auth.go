package docker

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/types"
	"github.com/docker/libcompose/docker"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

// ConfigAuthLookup will lookup registry auth info from cloud config
// if a context is set, it will also lookup auth info from the Docker config file
type ConfigAuthLookup struct {
	cfg                    *config.CloudConfig
	context                *docker.Context
	dockerConfigAuthLookup *docker.ConfigAuthLookup
}

func NewConfigAuthLookup(cfg *config.CloudConfig) *ConfigAuthLookup {
	return &ConfigAuthLookup{
		cfg: cfg,
	}
}

func populateRemaining(authConfig *types.AuthConfig) error {
	if authConfig.Auth == "" {
		return nil
	}

	decoded, err := base64.URLEncoding.DecodeString(authConfig.Auth)
	if err != nil {
		return err
	}

	decodedSplit := strings.Split(string(decoded), ":")
	if len(decodedSplit) != 2 {
		return fmt.Errorf("Invalid auth: %s", authConfig.Auth)
	}

	authConfig.Username = decodedSplit[0]
	authConfig.Password = decodedSplit[1]

	return nil
}

func (c *ConfigAuthLookup) SetConfig(cfg *config.CloudConfig) {
	c.cfg = cfg
}

func (c *ConfigAuthLookup) SetContext(context *docker.Context) {
	c.context = context
	c.dockerConfigAuthLookup = docker.NewConfigAuthLookup(context)
}

func (c *ConfigAuthLookup) Lookup(repoInfo *registry.RepositoryInfo) types.AuthConfig {
	if repoInfo == nil || repoInfo.Index == nil {
		return types.AuthConfig{}
	}
	authConfig := registry.ResolveAuthConfig(c.All(), repoInfo.Index)

	err := populateRemaining(&authConfig)
	if err != nil {
		log.Error(err)
		return types.AuthConfig{}
	}

	return authConfig
}

func (c *ConfigAuthLookup) All() map[string]types.AuthConfig {
	registryAuths := c.cfg.Rancher.RegistryAuths
	if c.dockerConfigAuthLookup != nil {
		for registry, authConfig := range c.dockerConfigAuthLookup.All() {
			registryAuths[registry] = authConfig
		}
	}
	return registryAuths
}
