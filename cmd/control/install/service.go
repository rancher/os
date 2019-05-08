package install

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
	"github.com/rancher/os/pkg/util/network"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
)

type ImageConfig struct {
	Image string `yaml:"image,omitempty"`
}

func GetCacheImageList(stage bool, cloudconfig, installType string, cfg *config.CloudConfig) []string {
	stageImages := make([]string, 0)
	if !stage || cloudconfig == "" || installType == "upgrade" {
		return stageImages
	}
	bytes, err := readConfigFile(cloudconfig)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to read cloud-config")
		return stageImages
	}
	r := make(map[interface{}]interface{})
	if err := yaml.Unmarshal(bytes, &r); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to unmarshal cloud-config")
		return stageImages
	}
	c := &config.CloudConfig{}
	if err := util.Convert(r, c); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to convert cloud-config")
		return stageImages
	}
	for key, value := range c.Rancher.ServicesInclude {
		if value {
			bytes, err = network.LoadServiceResource(key, true, cfg)
			if err != nil {
				log.WithFields(log.Fields{"err": err}).Fatal("Failed to load service resource")
				return stageImages
			}
			imageCfg := map[interface{}]ImageConfig{}
			if err := yaml.Unmarshal(bytes, &imageCfg); err != nil {
				log.WithFields(log.Fields{"err": err}).Fatal("Failed to unmarshal service")
				return stageImages
			}
			serviceImage := replaceRegistryDomain(imageCfg[key].Image)
			slice := strings.SplitN(serviceImage, "/", 2)
			if slice[0] == "${REGISTRY_DOMAIN}" {
				serviceImage = slice[1]
			}
			stageImages = append(stageImages, serviceImage)
		}
	}
	return stageImages
}

func readConfigFile(file string) ([]byte, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			content = []byte{}
		} else {
			return nil, err
		}
	}
	return content, err
}

func replaceRegistryDomain(image string) string {
	slice := strings.SplitN(image, "/", 2)
	if slice[0] == "${REGISTRY_DOMAIN}" {
		return slice[1]
	}
	return image
}
