package config

import (
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/xeipuuv/gojsonschema"
)

// TODO: use this function from libcompose
func convertKeysToStrings(item interface{}) interface{} {
	switch typedDatas := item.(type) {
	case map[string]interface{}:
		for key, value := range typedDatas {
			typedDatas[key] = convertKeysToStrings(value)
		}
		return typedDatas
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range typedDatas {
			stringKey := key.(string)
			newMap[stringKey] = convertKeysToStrings(value)
		}
		return newMap
	case []interface{}:
		for i, value := range typedDatas {
			typedDatas[i] = append(typedDatas, convertKeysToStrings(value))
		}
		return typedDatas
	default:
		return item
	}
}

func Validate(bytes []byte) (*gojsonschema.Result, error) {
	var rawCfg map[string]interface{}
	if err := yaml.Unmarshal([]byte(bytes), &rawCfg); err != nil {
		return nil, err
	}
	rawCfg = convertKeysToStrings(rawCfg).(map[string]interface{})
	loader := gojsonschema.NewGoLoader(rawCfg)
	schemaLoader := gojsonschema.NewStringLoader(schema)
	return gojsonschema.Validate(schemaLoader, loader)
}
