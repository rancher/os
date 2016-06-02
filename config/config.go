package config

import (
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/util"
)

func Merge(bytes []byte) error {
	data, err := readConfig(bytes, false)
	if err != nil {
		return err
	}
	existing, err := readConfig(nil, false, CloudConfigFile)
	if err != nil {
		return err
	}
	return WriteToFile(util.Merge(existing, data), CloudConfigFile)
}

func Export(private, full bool) (string, error) {
	rawCfg, err := LoadRawConfig(full)
	if !private {
		rawCfg = filterPrivateKeys(rawCfg)
	}

	bytes, err := yaml.Marshal(rawCfg)
	return string(bytes), err
}

func Get(key string) (interface{}, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	data := map[interface{}]interface{}{}
	if err := util.ConvertIgnoreOmitEmpty(cfg, &data); err != nil {
		return nil, err
	}

	v, _ := getOrSetVal(key, data, nil)
	return v, nil
}

func Set(key string, value interface{}) error {
	data := map[interface{}]interface{}{}
	_, data = getOrSetVal(key, data, value)

	existing, err := readConfig(nil, false, CloudConfigFile)
	if err != nil {
		return err
	}

	return WriteToFile(util.Merge(existing, data), CloudConfigFile)
}
