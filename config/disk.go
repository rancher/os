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
	private, config := filterDottedKeys(data, []string{
		"rancher.ssh",
		"rancher.docker.ca_key",
		"rancher.docker.ca_cert",
		"rancher.docker.server_key",
		"rancher.docker.server_cert",
	})

	err := writeToFile(config, LocalConfigFile)
	if err != nil {
		return err
	}

	return writeToFile(private, PrivateConfigFile)
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

		left = util.MapsUnion(left, right, util.Replace)
	}

	if bytes != nil && len(bytes) > 0 {
		right := make(map[interface{}]interface{})
		if err := yaml.Unmarshal(bytes, &right); err != nil {
			return nil, err
		}

		left = util.MapsUnion(left, right, util.Replace)
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
