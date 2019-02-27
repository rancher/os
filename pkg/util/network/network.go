package network

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	composeConfig "github.com/docker/libcompose/config"
)

var (
	ErrNoNetwork = errors.New("Networking not available to load resource")
	ErrNotFound  = errors.New("Failed to find resource")
)

func GetServices(urls []string) ([]string, error) {
	return getServices(urls, "services")
}

func GetConsoles(urls []string) ([]string, error) {
	return getServices(urls, "consoles")
}

func GetEngines(urls []string) ([]string, error) {
	return getServices(urls, "engines")
}

func getServices(urls []string, key string) ([]string, error) {
	result := []string{}

	for _, url := range urls {
		indexURL := fmt.Sprintf("%s/index.yml", url)
		content, err := LoadResource(indexURL, true)
		if err != nil {
			log.Errorf("Failed to load %s: %v", indexURL, err)
			continue
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			log.Errorf("Failed to unmarshal %s: %v", indexURL, err)
			continue
		}

		if list, ok := services[key]; ok {
			result = append(result, list...)
		}
	}

	return result, nil
}

func SetProxyEnvironmentVariables() {
	cfg := config.LoadConfig()
	if cfg.Rancher.Network.HTTPProxy != "" {
		err := os.Setenv("HTTP_PROXY", cfg.Rancher.Network.HTTPProxy)
		if err != nil {
			log.Errorf("Unable to set HTTP_PROXY: %s", err)
		}
	}
	if cfg.Rancher.Network.HTTPSProxy != "" {
		err := os.Setenv("HTTPS_PROXY", cfg.Rancher.Network.HTTPSProxy)
		if err != nil {
			log.Errorf("Unable to set HTTPS_PROXY: %s", err)
		}
	}
	if cfg.Rancher.Network.NoProxy != "" {
		err := os.Setenv("NO_PROXY", cfg.Rancher.Network.NoProxy)
		if err != nil {
			log.Errorf("Unable to set NO_PROXY: %s", err)
		}
	}
	if cfg.Rancher.Network.HTTPProxy != "" {
		config.Set("rancher.environment.http_proxy", cfg.Rancher.Network.HTTPProxy)
		config.Set("rancher.environment.HTTP_PROXY", cfg.Rancher.Network.HTTPProxy)
	}
	if cfg.Rancher.Network.HTTPSProxy != "" {
		config.Set("rancher.environment.https_proxy", cfg.Rancher.Network.HTTPSProxy)
		config.Set("rancher.environment.HTTPS_PROXY", cfg.Rancher.Network.HTTPSProxy)
	}
	if cfg.Rancher.Network.NoProxy != "" {
		config.Set("rancher.environment.no_proxy", cfg.Rancher.Network.NoProxy)
		config.Set("rancher.environment.NO_PROXY", cfg.Rancher.Network.NoProxy)
	}
}

func LoadFromNetworkWithCache(location string) ([]byte, error) {
	bytes := cacheLookup(location)
	if bytes != nil {
		return bytes, nil
	}
	return LoadFromNetwork(location)
}

func LoadFromNetwork(location string) ([]byte, error) {
	SetProxyEnvironmentVariables()

	var err error

	var resp *http.Response
	log.Debugf("LoadFromNetwork(%s)", location)
	resp, err = http.Get(location)
	log.Debugf("LoadFromNetwork(%s) returned %v, %v", location, resp, err)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 http response: %d", resp.StatusCode)
		}

		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		cacheAdd(location, bytes)
		return bytes, nil
	}

	return nil, err
}

func LoadResource(location string, network bool) ([]byte, error) {
	if strings.HasPrefix(location, "http:/") || strings.HasPrefix(location, "https:/") {
		if !network {
			return nil, ErrNoNetwork
		}
		return LoadFromNetworkWithCache(location)
	} else if strings.HasPrefix(location, "/") {
		return ioutil.ReadFile(location)
	}

	return nil, ErrNotFound
}

func serviceURL(url, name string) string {
	return fmt.Sprintf("%s/%s/%s.yml", url, name[0:1], name)
}

func LoadServiceResource(name string, useNetwork bool, cfg *config.CloudConfig) ([]byte, error) {
	bytes, err := LoadResource(name, useNetwork)
	if err == nil {
		log.Debugf("Loaded %s from %s", name, name)
		return bytes, nil
	}
	if err == ErrNoNetwork || !useNetwork {
		return nil, ErrNoNetwork
	}

	urls := cfg.Rancher.Repositories.ToArray()
	for _, url := range urls {
		serviceURL := serviceURL(url, name)
		bytes, err = LoadResource(serviceURL, useNetwork)
		if err == nil {
			log.Debugf("Loaded %s from %s", name, serviceURL)
			return bytes, nil
		}
	}

	return nil, err
}

func LoadMultiEngineResource(name string) ([]byte, error) {
	composeConfigs := map[string]composeConfig.ServiceConfigV1{}
	if _, err := os.Stat(config.MultiDockerConfFile); err == nil {
		multiEngineBytes, err := ioutil.ReadFile(config.MultiDockerConfFile)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(multiEngineBytes, &composeConfigs)
		if err != nil {
			return nil, err
		}
	}

	if _, ok := composeConfigs[name]; !ok {
		return nil, errors.New("Failed to found " + name + " from " + config.MultiDockerConfFile + " will load from network")
	}

	foundServiceConfig := map[string]composeConfig.ServiceConfigV1{}
	foundServiceConfig[name] = composeConfigs[name]
	bytes, err := yaml.Marshal(foundServiceConfig)
	if err == nil {
		return bytes, err
	}

	return nil, err
}

func UpdateCaches(urls []string, key string) error {
	for _, url := range urls {
		indexURL := fmt.Sprintf("%s/index.yml", url)
		content, err := UpdateCache(indexURL)
		if err != nil {
			return err
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			return err
		}

		list := services[key]
		for _, name := range list {
			serviceURL := serviceURL(url, name)
			// no need to handle error
			UpdateCache(serviceURL)
		}
	}
	return nil
}

func UpdateCache(location string) ([]byte, error) {
	if err := cacheRemove(location); err != nil {
		return []byte{}, err
	}
	content, err := LoadResource(location, true)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}
