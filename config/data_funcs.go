package config

import (
	log "github.com/Sirupsen/logrus"

	"regexp"
	"strconv"
	"strings"
)

func getOrSetVal(args string, data map[interface{}]interface{}, value interface{}) interface{} {
	parts := strings.Split(args, ".")

	for i, part := range parts {
		val, ok := data[part]
		last := i+1 == len(parts)

		// Reached end, set the value
		if last && value != nil {
			if s, ok := value.(string); ok {
				value = DummyMarshall(s)
			}

			data[part] = value
			return value
		}

		// Missing intermediate key, create key
		if !last && value != nil && !ok {
			newData := map[interface{}]interface{}{}
			data[part] = newData
			data = newData
			continue
		}

		if !ok {
			break
		}

		if last {
			return val
		}

		newData, ok := val.(map[interface{}]interface{})
		if !ok {
			break
		}

		data = newData
	}

	return ""
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
		keys := strings.Split(kv[0], ".")[1:]
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
