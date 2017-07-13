package network

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
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

func SetProxyEnvironmentVariables(cfg *config.CloudConfig) {
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
}

func loadFromNetwork(location string) ([]byte, error) {
	bytes := cacheLookup(location)
	if bytes != nil {
		return bytes, nil
	}

	cfg := config.LoadConfig()
	SetProxyEnvironmentVariables(cfg)

	var err error
	// Sven thinks that the dhcpcd --wait we added makes this less necessary
	//for i := 0; i < 300; i++ {
	updateDNSCache()

	var resp *http.Response
	log.Debugf("LoadFromNetwork(%s)", location)
	resp, err = http.Get(location)
	log.Debugf("LoadFromNetwork(%s) returned %v", location, resp)
	log.Debugf("LoadFromNetwork(%s) error %v", location, err)
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

	//	time.Sleep(100 * time.Millisecond)
	//}

	return nil, err
}

func LoadResource(location string, network bool) ([]byte, error) {
	if strings.HasPrefix(location, "http:/") || strings.HasPrefix(location, "https:/") {
		if !network {
			return nil, ErrNoNetwork
		}
		return loadFromNetwork(location)
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
