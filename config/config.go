package config

import (
	"fmt"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/util"
)

var keysToStringify = []string{
	"command",
	"dns",
	"dns_search",
	"entrypoint",
	"env_file",
	"environment",
	"labels",
	"links",
}

func isPathToStringify(path []interface{}) bool {
	l := len(path)
	if l == 0 {
		return false
	}
	if sk, ok := path[l-1].(string); ok {
		return util.Contains(keysToStringify, sk)
	}
	return false
}

func stringifyValue(data interface{}, path []interface{}) interface{} {
	switch data := data.(type) {
	case map[interface{}]interface{}:
		result := make(map[interface{}]interface{}, len(data))
		if isPathToStringify(path) {
			for k, v := range data {
				switch v := v.(type) {
				case []interface{}:
					result[k] = stringifyValue(v, append(path, k))
				case map[interface{}]interface{}:
					result[k] = stringifyValue(v, append(path, k))
				default:
					result[k] = fmt.Sprint(v)
				}
			}
		} else {
			for k, v := range data {
				result[k] = stringifyValue(v, append(path, k))
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(data))
		if isPathToStringify(path) {
			for k, v := range data {
				result[k] = fmt.Sprint(v)
			}
		} else {
			for k, v := range data {
				result[k] = stringifyValue(v, append(path, k))
			}
		}
		return result
	default:
		return data
	}
}

func StringifyValues(data map[interface{}]interface{}) map[interface{}]interface{} {
	return stringifyValue(data, nil).(map[interface{}]interface{})
}

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
	rawCfg := loadRawDiskConfig(full)
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

func Set(key string, value interface{}) error {
	existing, err := readConfigs(nil, false, true, CloudConfigFile)
	if err != nil {
		return err
	}

	_, modified := getOrSetVal(key, existing, value)
	return WriteToFile(modified, CloudConfigFile)
}
