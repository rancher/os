package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	values "github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/data/convert"
	schemas2 "github.com/rancher/wrangler/pkg/schemas"
	"sigs.k8s.io/yaml"
)

var (
	defaultMappers = schemas2.Mappers{
		NewToMap(),
		NewToSlice(),
		NewToBool(),
		&FuzzyNames{},
	}
	schemas = schemas2.EmptySchemas().Init(func(s *schemas2.Schemas) *schemas2.Schemas {
		s.AddMapper("config", defaultMappers)
		s.AddMapper("rancherOS", defaultMappers)
		s.AddMapper("install", defaultMappers)
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
		keyName = strings.ReplaceAll(keyName, "RANCHEROS_", "COS_")
		if data, ok := v.(map[string]interface{}); ok {
			subResult := mapToEnv(keyName+"_", data)
			result = append(result, subResult...)
		} else {
			result = append(result, fmt.Sprintf("%s=%v", keyName, v))
		}
	}
	return result
}

func readFileFunc(path string) func() (map[string]interface{}, error) {
	return func() (map[string]interface{}, error) {
		return readFile(path)
	}
}

func readNested(data map[string]interface{}) (map[string]interface{}, error) {
	var (
		nestedConfigFiles = convert.ToStringSlice(values.GetValueN(data, "rancheros", "install", "configUrl"))
		funcs             []reader
	)

	for _, nestedConfigFile := range nestedConfigFiles {
		funcs = append(funcs, readFileFunc(nestedConfigFile))
	}

	funcs = append(funcs, func() (map[string]interface{}, error) {
		return data, nil
	})

	return merge(funcs...)
}

func readFile(path string) (result map[string]interface{}, _ error) {
	result = map[string]interface{}{}
	defer func() {
		if v, ok := result["install"]; ok {
			values.PutValue(result, v, "rancheros", "install")
		}
	}()

	switch {
	case strings.HasPrefix(path, "http://"):
		fallthrough
	case strings.HasPrefix(path, "https://"):
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		buffer, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}

		return result, yaml.Unmarshal(buffer, &result)
	case strings.HasPrefix(path, "tftp://"):
		return tftpGet(path)
	}

	f, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(f, &data); err != nil {
		return nil, err
	}

	return readNested(data)
}

type reader func() (map[string]interface{}, error)

func merge(readers ...reader) (map[string]interface{}, error) {
	d := map[string]interface{}{}
	for _, r := range readers {
		newData, err := r()
		if err != nil {
			return nil, err
		}
		if err := schema.Mapper.ToInternal(newData); err != nil {
			return nil, err
		}
		d = values.MergeMapsConcatSlice(d, newData)
	}
	return d, nil
}

func readConfigMap(cfg string) (map[string]interface{}, error) {
	data, err := merge(readCmdline, readFileFunc(cfg))
	if err != nil {
		return nil, err
	}
	if cfg != "" {
		values.PutValue(data, cfg, "rancheros", "install", "configUrl")
	}
	return data, nil
}

func ToFile(cfg Config, output string) error {
	data, err := ToBytes(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(output, data, 0600)
}

func ToBytes(cfg Config) ([]byte, error) {
	data, err := merge(readFileFunc(cfg.RancherOS.Install.ConfigURL), func() (map[string]interface{}, error) {
		return convert.EncodeToMap(cfg)
	})
	if err != nil {
		return nil, err
	}
	values.RemoveValue(data, "install")
	values.RemoveValue(data, "rancheros", "install")
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	return append([]byte("#cloud-config\n"), bytes...), nil
}

func ReadConfig(cfg string) (result Config, err error) {
	data, err := readConfigMap(cfg)
	if err != nil {
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

	procCmdLine := os.Getenv("PROC_CMDLINE")
	if procCmdLine == "" {
		procCmdLine = "/proc/cmdline"
	}
	bytes, err := ioutil.ReadFile(procCmdLine)
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

	return readNested(data)
}
