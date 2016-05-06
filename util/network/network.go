package network

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/config"
)

var (
	ErrNoNetwork = errors.New("Networking not available to load resource")
	ErrNotFound  = errors.New("Failed to find resource")
)

func GetServices(urls []string) ([]string, error) {
	result := []string{}

	for _, url := range urls {
		indexUrl := fmt.Sprintf("%s/index.yml", url)
		content, err := LoadResource(indexUrl, true, []string{})
		if err != nil {
			log.Errorf("Failed to load %s: %v", indexUrl, err)
			continue
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			log.Errorf("Failed to unmarshal %s: %v", indexUrl, err)
			continue
		}

		if list, ok := services["services"]; ok {
			result = append(result, list...)
		}
	}

	return result, nil
}

func SetProxyEnvironmentVariables(cfg *config.CloudConfig) {
	if cfg.Rancher.Network.HttpProxy != "" {
		err := os.Setenv("HTTP_PROXY", cfg.Rancher.Network.HttpProxy)
		if err != nil {
			log.Errorf("Unable to set HTTP_PROXY: %s", err)
		}
	}
	if cfg.Rancher.Network.HttpsProxy != "" {
		err := os.Setenv("HTTPS_PROXY", cfg.Rancher.Network.HttpsProxy)
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

func retryHttp(f func() (*http.Response, error), times int) (resp *http.Response, err error) {
	for i := 0; i < times; i++ {
		if resp, err = f(); err == nil {
			return
		}
		log.Warnf("Error making HTTP request: %s. Retrying", err)
	}
	return
}

func LoadResource(location string, network bool, urls []string) ([]byte, error) {
	var bytes []byte
	err := ErrNotFound

	if strings.HasPrefix(location, "http:/") || strings.HasPrefix(location, "https:/") {
		if !network {
			return nil, ErrNoNetwork
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return nil, err
		}

		SetProxyEnvironmentVariables(cfg)

		resp, err := retryHttp(func() (*http.Response, error) {
			return http.Get(location)
		}, 8)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 http response: %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	} else if strings.HasPrefix(location, "/") {
		return ioutil.ReadFile(location)
	} else if len(location) > 0 {
		for _, url := range urls {
			ymlUrl := fmt.Sprintf("%s/%s/%s.yml", url, location[0:1], location)
			bytes, err = LoadResource(ymlUrl, network, []string{})
			if err == nil {
				log.Debugf("Loaded %s from %s", location, ymlUrl)
				return bytes, nil
			}
		}
	}

	return nil, err
}
