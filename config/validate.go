package config

import (
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/xeipuuv/gojsonschema"
)

// TODO: use this function from libcompose
func ConvertKeysToStrings(item interface{}) interface{} {
	switch typedDatas := item.(type) {
	case map[string]interface{}:
		for key, value := range typedDatas {
			typedDatas[key] = ConvertKeysToStrings(value)
		}
		return typedDatas
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range typedDatas {
			stringKey := key.(string)
			newMap[stringKey] = ConvertKeysToStrings(value)
		}
		return newMap
	case []interface{}:
		for i, value := range typedDatas {
			typedDatas[i] = ConvertKeysToStrings(value)
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
	rawCfg = ConvertKeysToStrings(rawCfg).(map[string]interface{})
	loader := gojsonschema.NewGoLoader(rawCfg)
	schemaLoader := gojsonschema.NewStringLoader(schema)
	return gojsonschema.Validate(schemaLoader, loader)
}
