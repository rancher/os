package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	values "github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/data/convert"
	schemas2 "github.com/rancher/wrangler/pkg/schemas"
)

var (
	schemas = schemas2.EmptySchemas().Init(func(s *schemas2.Schemas) *schemas2.Schemas {
		s.DefaultMapper = func() schemas2.Mapper {
			return schemas2.Mappers{
				NewToMap(),
				NewToSlice(),
				NewToBool(),
				&FuzzyNames{},
			}
		}
		return s
	}).MustImport(Config{})
	schema = schemas.Schema("config")
)

func ToEnv(cfg Config) ([]string, error) {
	data, err := convert.EncodeToMap(&cfg)
	if err != nil {
		return nil, err
	}

	return mapToEnv("", data), nil
}

func mapToEnv(prefix string, data map[string]interface{}) []string {
	var result []string
	for k, v := range data {
		keyName := strings.ToUpper(prefix + convert.ToYAMLKey(k))
		keyName = strings.ReplaceAll(keyName, "ELEMENTAL_", "COS_")
		if data, ok := v.(map[string]interface{}); ok {
			subResult := mapToEnv(keyName+"_", data)
			result = append(result, subResult...)
		} else {
			result = append(result, fmt.Sprintf("%s=%v", keyName, v))
		}
	}
	return result
}

func ReadConfig() (Config, error) {
	result := Config{}

	data, err := readCmdline()
	if err != nil {
		return result, err
	}
	if err := schema.Mapper.ToInternal(data); err != nil {
		return result, err
	}

	return result, convert.ToObj(data, &result)
}

func readCmdline() (map[string]interface{}, error) {
	//supporting regex https://regexr.com/4mq0s
	parser, err := regexp.Compile(`(\"[^\"]+\")|([^\s]+=(\"[^\"]+\")|([^\s]+))`)
	if err != nil {
		return nil, nil
	}

	bytes, err := ioutil.ReadFile("/proc/cmdline")
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	for _, item := range parser.FindAllString(string(bytes), -1) {
		parts := strings.SplitN(item, "=", 2)
		value := "true"
		if len(parts) > 1 {
			value = strings.Trim(parts[1], `"`)
		}
		keys := strings.Split(strings.Trim(parts[0], `"`), ".")
		existing, ok := values.GetValue(data, keys...)
		if ok {
			switch v := existing.(type) {
			case string:
				values.PutValue(data, []string{v, value}, keys...)
			case []string:
				values.PutValue(data, append(v, value), keys...)
			}
		} else {
			values.PutValue(data, value, keys...)
		}
	}

	return data, nil
}
