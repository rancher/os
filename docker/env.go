package docker

import (
	"fmt"
	"strings"

	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

type ConfigEnvironment struct {
	cfg *config.CloudConfig
}

func NewConfigEnvironment(cfg *config.CloudConfig) *ConfigEnvironment {
	return &ConfigEnvironment{
		cfg: cfg,
	}
}

func appendEnv(array []string, key, value string) []string {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		key = parts[1]
	}

	return append(array, fmt.Sprintf("%s=%s", key, value))
}

func environmentFromCloudConfig(cfg *config.CloudConfig) map[string]string {
	environment := cfg.Rancher.Environment
	if cfg.Rancher.Network.HTTPProxy != "" {
		environment["http_proxy"] = cfg.Rancher.Network.HTTPProxy
		environment["HTTP_PROXY"] = cfg.Rancher.Network.HTTPProxy
	}
	if cfg.Rancher.Network.HTTPSProxy != "" {
		environment["https_proxy"] = cfg.Rancher.Network.HTTPSProxy
		environment["HTTPS_PROXY"] = cfg.Rancher.Network.HTTPSProxy
	}
	if cfg.Rancher.Network.NoProxy != "" {
		environment["no_proxy"] = cfg.Rancher.Network.NoProxy
		environment["NO_PROXY"] = cfg.Rancher.Network.NoProxy
	}
	if v := config.GetKernelVersion(); v != "" {
		environment["KERNEL_VERSION"] = v
		log.Debugf("Using /proc/version to set rancher.environment.KERNEL_VERSION = %s", v)
	}
	return environment
}

func lookupKeys(cfg *config.CloudConfig, keys ...string) []string {
	environment := environmentFromCloudConfig(cfg)

	for _, key := range keys {
		if strings.HasSuffix(key, "*") {
			result := []string{}
			for envKey, envValue := range environment {
				keyPrefix := key[:len(key)-1]
				if strings.HasPrefix(envKey, keyPrefix) {
					result = appendEnv(result, envKey, envValue)
				}
			}

			if len(result) > 0 {
				return result
			}
		} else if value, ok := environment[key]; ok {
			return appendEnv([]string{}, key, value)
		}
	}

	return []string{}
}

func (c *ConfigEnvironment) SetConfig(cfg *config.CloudConfig) {
	c.cfg = cfg
}

func (c *ConfigEnvironment) Lookup(key, serviceName string, serviceConfig *composeConfig.ServiceConfig) []string {
	fullKey := fmt.Sprintf("%s/%s", serviceName, key)
	return lookupKeys(c.cfg, fullKey, key)
}
