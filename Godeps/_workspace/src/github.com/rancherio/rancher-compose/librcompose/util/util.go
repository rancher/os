package util

import "gopkg.in/yaml.v2"

func Convert(src, target interface{}) error {
	newBytes, err := yaml.Marshal(src)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(newBytes, target)
}
