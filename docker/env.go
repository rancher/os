package docker

import (
	"fmt"
	"strings"

	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config"
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
	if cfg.Rancher.Network.HttpProxy != "" {
		environment["http_proxy"] = cfg.Rancher.Network.HttpProxy
		environment["HTTP_PROXY"] = cfg.Rancher.Network.HttpProxy
	}
	if cfg.Rancher.Network.HttpsProxy != "" {
		environment["https_proxy"] = cfg.Rancher.Network.HttpsProxy
		environment["HTTPS_PROXY"] = cfg.Rancher.Network.HttpsProxy
	}
	if cfg.Rancher.Network.NoProxy != "" {
		environment["no_proxy"] = cfg.Rancher.Network.NoProxy
		environment["NO_PROXY"] = cfg.Rancher.Network.NoProxy
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
