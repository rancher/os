package cmdline

import (
	"io/ioutil"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/pkg/util"
)

func Read(parseAll bool) (m map[interface{}]interface{}, err error) {
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return nil, err
	}

	if len(cmdLine) == 0 {
		return nil, nil
	}

	cmdLineObj := Parse(strings.TrimSpace(util.UnescapeKernelParams(string(cmdLine))), parseAll)

	return cmdLineObj, nil
}

func GetCmdline(key string) interface{} {
	parseAll := true
	if strings.HasPrefix(key, "cc.") || strings.HasPrefix(key, "rancher.") {
		// the normal case
		parseAll = false
	}
	cmdline, _ := Read(parseAll)
	v, _ := GetOrSetVal(key, cmdline, nil)
	return v
}

func GetOrSetVal(args string, data map[interface{}]interface{}, value interface{}) (interface{}, map[interface{}]interface{}) {
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
				value = UnmarshalOrReturnString(s)
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

func UnmarshalOrReturnString(value string) (result interface{}) {
	value = strings.Replace(value, "\n", newlineMagicString, -1)
	value = strings.Replace(value, ":", colonMagicString, -1)
	value = strings.Replace(value, "?", questionMarkMagicString, -1)
	if err := yaml.Unmarshal([]byte(value), &result); err != nil {
		result = value
	}
	result = reverseReplacement(result)
	return
}

func Parse(cmdLine string, parseAll bool) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}

outer:
	for _, part := range strings.Split(cmdLine, " ") {
		if strings.HasPrefix(part, "cc.") {
			part = part[3:]
		} else if !strings.HasPrefix(part, "rancher.") {
			if !parseAll {
				continue
			}
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
				current[key] = UnmarshalOrReturnString(value)
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

	return result
}
