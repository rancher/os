package config

import (
	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"strings"

	"github.com/rancher/os/util"
)

type CfgFunc func(*CloudConfig) (*CloudConfig, error)

func ChainCfgFuncs(cfg *CloudConfig, cfgFuncs ...CfgFunc) (*CloudConfig, error) {
	for i, cfgFunc := range cfgFuncs {
		log.Debugf("[%d/%d] Starting", i+1, len(cfgFuncs))
		var err error
		if cfg, err = cfgFunc(cfg); err != nil {
			log.Errorf("Failed [%d/%d] %d%%", i+1, len(cfgFuncs), ((i + 1) * 100 / len(cfgFuncs)))
			return cfg, err
		}
		log.Debugf("[%d/%d] Done %d%%", i+1, len(cfgFuncs), ((i + 1) * 100 / len(cfgFuncs)))
	}
	return cfg, nil
}

func filterKey(data map[interface{}]interface{}, key []string) (filtered, rest map[interface{}]interface{}) {
	if len(key) == 0 {
		return data, map[interface{}]interface{}{}
	}

	filtered = map[interface{}]interface{}{}
	rest = util.MapCopy(data)

	k := key[0]
	if d, ok := data[k]; ok {
		switch d := d.(type) {

		case map[interface{}]interface{}:
			f, r := filterKey(d, key[1:])

			if len(f) != 0 {
				filtered[k] = f
			}

			if len(r) != 0 {
				rest[k] = r
			} else {
				delete(rest, k)
			}

		default:
			filtered[k] = d
			delete(rest, k)
		}

	}

	return
}

func filterPrivateKeys(data map[interface{}]interface{}) map[interface{}]interface{} {
	for _, privateKey := range PrivateKeys {
		_, data = filterKey(data, strings.Split(privateKey, "."))
	}

	return data
}

func getOrSetVal(args string, data map[interface{}]interface{}, value interface{}) (interface{}, map[interface{}]interface{}) {
	parts := strings.Split(args, ".")

	tData := data
	if value != nil {
		tData = util.MapCopy(data)
	}
	t := tData
	for i, part := range parts {
		val, ok := t[part]
		last := i+1 == len(parts)

		// Reached end, set the value
		if last && value != nil {
			if s, ok := value.(string); ok {
				value = unmarshalOrReturnString(s)
			}

			t[part] = value
			return value, tData
		}

		// Missing intermediate key, create key
		if !last && value != nil && !ok {
			newData := map[interface{}]interface{}{}
			t[part] = newData
			t = newData
			continue
		}

		if !ok {
			break
		}

		if last {
			return val, tData
		}

		newData, ok := val.(map[interface{}]interface{})
		if !ok {
			break
		}

		t = newData
	}

	return "", tData
}

// Replace newlines and colons with random strings
// This is done to avoid YAML treating these as special characters
var (
	newlineMagicString = "9XsJcx6dR5EERYCC"
	colonMagicString   = "V0Rc21pIVknMm2rr"
)

func reverseReplacement(result interface{}) interface{} {
	switch val := result.(type) {
	case map[interface{}]interface{}:
		for k, v := range val {
			val[k] = reverseReplacement(v)
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = reverseReplacement(item)
		}
		return val
	case string:
		val = strings.Replace(val, newlineMagicString, "\n", -1)
		val = strings.Replace(val, colonMagicString, ":", -1)
		return val
	}

	return result
}

func unmarshalOrReturnString(value string) (result interface{}) {
	value = strings.Replace(value, "\n", newlineMagicString, -1)
	value = strings.Replace(value, ":", colonMagicString, -1)
	if err := yaml.Unmarshal([]byte(value), &result); err != nil {
		result = value
	}
	result = reverseReplacement(result)
	return
}

func parseCmdline(cmdLine string) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})

outer:
	for _, part := range strings.Split(cmdLine, " ") {
		if !strings.HasPrefix(part, "rancher.") {
			continue
		}

		var value string
		kv := strings.SplitN(part, "=", 2)

		if len(kv) == 1 {
			value = "true"
		} else {
			value = kv[1]
		}

		current := result
		keys := strings.Split(kv[0], ".")
		for i, key := range keys {
			if i == len(keys)-1 {
				current[key] = unmarshalOrReturnString(value)
			} else {
				if obj, ok := current[key]; ok {
					if newCurrent, ok := obj.(map[interface{}]interface{}); ok {
						current = newCurrent
					} else {
						continue outer
					}
				} else {
					newCurrent := make(map[interface{}]interface{})
					current[key] = newCurrent
					current = newCurrent
				}
			}
		}
	}

	log.Debugf("Input obj %v", result)
	return result
}
