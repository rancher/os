package config

import (
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/util"
)

func Merge(bytes []byte) error {
	data, err := readConfigs(bytes, false, true)
	if err != nil {
		return err
	}
	existing, err := readConfigs(nil, false, true, CloudConfigFile)
	if err != nil {
		return err
	}
	return WriteToFile(util.Merge(existing, data), CloudConfigFile)
}

func Export(private, full bool) (string, error) {
	rawCfg := loadRawDiskConfig("", full)
	if !private {
		rawCfg = filterPrivateKeys(rawCfg)
	}

	bytes, err := yaml.Marshal(rawCfg)
	return string(bytes), err
}

func Get(key string) (interface{}, error) {
	cfg := LoadConfig()

	data := map[interface{}]interface{}{}
	if err := util.ConvertIgnoreOmitEmpty(cfg, &data); err != nil {
		return nil, err
	}

	v, _ := getOrSetVal(key, data, nil)
	return v, nil
}

func GetCmdline(key string) interface{} {
	cmdline := readCmdline()
	v, _ := getOrSetVal(key, cmdline, nil)
	return v
}

func Set(key string, value interface{}) error {
	existing, err := readConfigs(nil, false, true, CloudConfigFile)
	if err != nil {
		return err
	}

	_, modified := getOrSetVal(key, existing, value)

	c := &CloudConfig{}
	if err = util.Convert(modified, c); err != nil {
		return err
	}

	return WriteToFile(modified, CloudConfigFile)
}
