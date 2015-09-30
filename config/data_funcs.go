package config

import (
	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/util"
	"regexp"
	"strconv"
	"strings"
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

func filterDottedKeys(data map[interface{}]interface{}, keys []string) (filtered, rest map[interface{}]interface{}) {
	filtered = map[interface{}]interface{}{}
	rest = util.MapCopy(data)

	for _, key := range keys {
		f, r := filterKey(data, strings.Split(key, "."))
		filtered = util.MapsUnion(filtered, f)
		rest = util.MapsIntersection(rest, r)
	}

	return
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
				value = DummyMarshall(s)
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

func DummyMarshall(value string) interface{} {
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		result := []string{}
		for _, i := range strings.Split(value[1:len(value)-1], ",") {
			result = append(result, strings.TrimSpace(i))
		}
		return result
	}

	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else if ok, _ := regexp.MatchString("^[0-9]+$", value); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		return i
	}

	return value
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
				current[key] = DummyMarshall(value)
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
