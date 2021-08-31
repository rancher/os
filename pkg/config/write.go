package config

import (
	"github.com/rancher/wrangler/pkg/data/convert"
	"sigs.k8s.io/yaml"
)

func PrintInstall(cfg Config) ([]byte, error) {
	if cfg.Elemental.Install.Password != "" {
		cfg.Elemental.Install.Password = "******"
	}
	data, err := convert.EncodeToMap(cfg.Elemental.Install)
	if err != nil {
		return nil, err
	}

	toYAMLKeys(data)
	return yaml.Marshal(data)
}

func toYAMLKeys(data map[string]interface{}) {
	for k, v := range data {
		if sub, ok := v.(map[string]interface{}); ok {
			toYAMLKeys(sub)
		}
		newK := convert.ToYAMLKey(k)
		if newK != k {
			delete(data, k)
			data[newK] = v
		}
	}
}
