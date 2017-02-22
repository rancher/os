package config

import (
	"regexp"
	"strconv"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/log"

	"strings"

	"github.com/fatih/structs"
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
				value = dummyUnmarshall(s)
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

	return nil, tData
}

func checkTypeAndSetVal(args string, data map[interface{}]interface{}, value interface{}) map[interface{}]interface{} {
	if getFieldType(args, &CloudConfig{}) == "slice" {
		prevValue, _ := getOrSetVal(args, data, nil)
		prevSlice, ok := prevValue.([]interface{})
		var newVal []interface{}
		if ok {
			newVal = prevSlice
		}
		switch v := value.(type) {
		case []interface{}:
			newVal = append(newVal, v...)
		case []string:
			for _, s := range v {
				newVal = append(newVal, s)
			}
		case string:
			newVal = append(newVal, v)
		}
		_, data = getOrSetVal(args, data, newVal)
	} else {
		_, data = getOrSetVal(args, data, value)
	}
	return data
}

// Replace newlines, colons, and question marks with random strings
// This is done to avoid YAML treating these as special characters
var (
	newlineMagicString      = "9XsJcx6dR5EERYCC"
	colonMagicString        = "V0Rc21pIVknMm2rr"
	questionMarkMagicString = "FoPL6JLMAaJqKMJT"
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
		val = strings.Replace(val, questionMarkMagicString, "?", -1)
		return val
	}

	return result
}

func unmarshalOrReturnString(value string) (result interface{}) {
	value = strings.Replace(value, "\n", newlineMagicString, -1)
	value = strings.Replace(value, ":", colonMagicString, -1)
	value = strings.Replace(value, "?", questionMarkMagicString, -1)
	if err := yaml.Unmarshal([]byte(value), &result); err != nil {
		result = value
	}
	result = reverseReplacement(result)
	return
}

func dummyUnmarshall(value string) interface{} {
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else if ok, _ := regexp.MatchString("^[0-9]+$", value); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return i
		}
	}
	return value
}

type Field interface {
	Fields() []*structs.Field
	FieldOk(string) (*structs.Field, bool)
}

func getFieldNameByTag(field Field, name string) *structs.Field {
	for _, currentField := range field.Fields() {
		currentName := currentField.Name()
		tagSplit := strings.Split(currentField.Tag("yaml"), ",")
		if len(tagSplit) > 0 && tagSplit[0] != "" {
			currentName = tagSplit[0]
		}
		if currentName == name {
			finalField, ok := field.FieldOk(currentField.Name())
			if !ok {
				return nil
			}
			return finalField
		}
	}
	return nil
}

// Too slow
func getFieldType(args string, obj interface{}) string {
	fields := strings.Split(args, ".")

	if len(fields) < 1 {
		return ""
	}

	s := structs.New(obj)
	currentField := getFieldNameByTag(s, fields[0])
	if currentField == nil {
		return ""
	}

	for _, field := range fields[1:] {
		currentField = getFieldNameByTag(currentField, field)
		if currentField == nil {
			return ""
		}
	}

	return currentField.Kind().String()
}

func parseCmdline(cmdLine string) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})

	for _, part := range strings.Split(cmdLine, " ") {
		if strings.HasPrefix(part, "cc.") {
			part = part[3:]
		} else if !strings.HasPrefix(part, "rancher.") {
			continue
		}

		var value interface{}
		kv := strings.SplitN(part, "=", 2)

		if len(kv) == 1 {
			value = true
		} else {
			val := kv[1]
			if len(val) > 2 && val[0] == '[' && val[len(val)-1] == ']' {
				// Read legacy array format
				_, result = getOrSetVal(kv[0], result, unmarshalOrReturnString(val))
				continue
			}
			value = kv[1]
		}
		result = checkTypeAndSetVal(kv[0], result, value)
	}

	log.Debugf("Input obj %v", result)
	return result
}
