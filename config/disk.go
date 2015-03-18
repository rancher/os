package config

import (
	"io/ioutil"
	"os"

	"github.com/rancherio/os/util"
	"gopkg.in/yaml.v2"
)

func writeToFile(data interface{}, filename string) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, content, 400)
}

func saveToDisk(data map[interface{}]interface{}) error {
	config := make(map[interface{}]interface{})
	private := make(map[interface{}]interface{})

	for k, v := range data {
		if k == "ssh" {
			private[k] = v
		} else if k == "user_docker" {
			var userDockerConfig DockerConfig
			var userDockerConfigPrivate DockerConfig
			err := util.Convert(v, &userDockerConfig)
			if err != nil {
				return err
			}

			userDockerConfigPrivate.CAKey = userDockerConfig.CAKey
			userDockerConfigPrivate.CACert = userDockerConfig.CACert
			userDockerConfigPrivate.ServerKey = userDockerConfig.ServerKey
			userDockerConfigPrivate.ServerCert = userDockerConfig.ServerCert

			userDockerConfig.CAKey = ""
			userDockerConfig.CACert = ""
			userDockerConfig.ServerKey = ""
			userDockerConfig.ServerCert = ""

			config[k] = userDockerConfig
			private[k] = userDockerConfigPrivate
		} else {
			config[k] = v
		}
	}

	err := writeToFile(config, ConfigFile)
	if err != nil {
		return err
	}

	return writeToFile(private, PrivateConfigFile)
}

func readSavedConfig(bytes []byte) (map[interface{}]interface{}, error) {
	return readConfig(bytes, CloudConfigFile, ConfigFile, PrivateConfigFile)
}

func readConfig(bytes []byte, files ...string) (map[interface{}]interface{}, error) {
	// You can't just overlay yaml bytes on to maps, it won't merge, but instead
	// just override the keys and not merge the map values.
	left := make(map[interface{}]interface{})
	for _, conf := range files {
		content, err := readConfigFile(conf)
		if err != nil {
			return nil, err
		}

		right := make(map[interface{}]interface{})
		err = yaml.Unmarshal(content, &right)
		if err != nil {
			return nil, err
		}

		util.MergeMaps(left, right)
	}

	if bytes != nil && len(bytes) > 0 {
		right := make(map[interface{}]interface{})
		if err := yaml.Unmarshal(bytes, &right); err != nil {
			return nil, err
		}

		util.MergeMaps(left, right)
	}

	return left, nil
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
