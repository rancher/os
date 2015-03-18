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
	data := make(map[interface{}]interface{})
	for _, conf := range files {
		content, err := readConfigFile(conf)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(content, &data)
		if err != nil {
			return nil, err
		}
	}

	if bytes != nil && len(bytes) > 0 {
		if err := yaml.Unmarshal(bytes, &data); err != nil {
			return nil, err
		}
	}

	return data, nil
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
