package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/Sirupsen/logrus"

	"reflect"
)

var (
	ErrNoNetwork = errors.New("Networking not available to load resource")
	ErrNotFound  = errors.New("Failed to find resource")
)

type AnyMap map[interface{}]interface{}

func Contains(values []string, value string) bool {
	if len(value) == 0 {
		return false
	}

	for _, i := range values {
		if i == value {
			return true
		}
	}

	return false
}

type ReturnsErr func() error

func FileCopy(src, dest string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { err = in.Close() }()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { err = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return
}

func Convert(from, to interface{}) error {
	bytes, err := yaml.Marshal(from)
	if err != nil {
		log.WithFields(log.Fields{"from": from, "err": err}).Warn("Error serializing to YML")
		return err
	}

	return yaml.Unmarshal(bytes, to)
}

func Copy(d interface{}) interface{} {
	switch d := d.(type) {
	case map[interface{}]interface{}:
		return MapCopy(d)
	case []interface{}:
		return SliceCopy(d)
	default:
		return d
	}
}

func Replace(l, r interface{}) interface{} {
	return r
}

func Equal(l, r interface{}) interface{} {
	if reflect.DeepEqual(l, r) {
		return l
	}
	return nil
}

func Filter(xs []interface{}, p func(x interface{}) bool) []interface{} {
	return FlatMap(xs, func(x interface{}) []interface{} {
		if p(x) {
			return []interface{}{x}
		}
		return []interface{}{}
	})
}

func FilterStrings(xs []string, p func(x string) bool) []string {
	return FlatMapStrings(xs, func(x string) []string {
		if p(x) {
			return []string{x}
		}
		return []string{}
	})
}

func Map(xs []interface{}, f func(x interface{}) interface{}) []interface{} {
	return FlatMap(xs, func(x interface{}) []interface{} { return []interface{}{f(x)} })
}

func FlatMap(xs []interface{}, f func(x interface{}) []interface{}) []interface{} {
	result := []interface{}{}
	for _, x := range xs {
		result = append(result, f(x)...)
	}
	return result
}

func FlatMapStrings(xs []string, f func(x string) []string) []string {
	result := []string{}
	for _, x := range xs {
		result = append(result, f(x)...)
	}
	return result
}

func MapsUnion(left, right map[interface{}]interface{}) map[interface{}]interface{} {
	result := MapCopy(left)

	for k, r := range right {
		if l, ok := left[k]; ok {
			switch l := l.(type) {
			case map[interface{}]interface{}:
				switch r := r.(type) {
				case map[interface{}]interface{}:
					result[k] = MapsUnion(l, r)
				default:
					result[k] = Replace(l, r)
				}
			default:
				result[k] = Replace(l, r)
			}
		} else {
			result[k] = Copy(r)
		}
	}

	return result
}

func MapsDifference(left, right map[interface{}]interface{}) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}

	for k, l := range left {
		if r, ok := right[k]; ok {
			switch l := l.(type) {
			case map[interface{}]interface{}:
				switch r := r.(type) {
				case map[interface{}]interface{}:
					if len(l) == 0 && len(r) == 0 {
						continue
					} else if len(l) == 0 {
						result[k] = l
					} else if v := MapsDifference(l, r); len(v) > 0 {
						result[k] = v
					}
				default:
					if v := Equal(l, r); v == nil {
						result[k] = l
					}
				}
			default:
				if v := Equal(l, r); v == nil {
					result[k] = l
				}
			}
		} else {
			result[k] = l
		}
	}

	return result
}

func MapsIntersection(left, right map[interface{}]interface{}) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}

	for k, l := range left {
		if r, ok := right[k]; ok {
			switch l := l.(type) {
			case map[interface{}]interface{}:
				switch r := r.(type) {
				case map[interface{}]interface{}:
					result[k] = MapsIntersection(l, r)
				default:
					if v := Equal(l, r); v != nil {
						result[k] = v
					}
				}
			default:
				if v := Equal(l, r); v != nil {
					result[k] = v
				}
			}
		}
	}

	return result
}

func MapCopy(data map[interface{}]interface{}) map[interface{}]interface{} {
	result := map[interface{}]interface{}{}
	for k, v := range data {
		result[k] = Copy(v)
	}
	return result
}

func SliceCopy(data []interface{}) []interface{} {
	result := make([]interface{}, len(data), len(data))
	for k, v := range data {
		result[k] = Copy(v)
	}
	return result
}

func ToStrings(data []interface{}) []string {
	result := make([]string, len(data), len(data))
	for k, v := range data {
		result[k] = v.(string)
	}
	return result
}

func GetServices(urls []string) ([]string, error) {
	result := []string{}

	for _, url := range urls {
		indexUrl := fmt.Sprintf("%s/index.yml", url)
		content, err := LoadResource(indexUrl, true, []string{})
		if err != nil {
			log.Errorf("Failed to load %s: %v", indexUrl, err)
			continue
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			log.Errorf("Failed to unmarshal %s: %v", indexUrl, err)
			continue
		}

		if list, ok := services["services"]; ok {
			result = append(result, list...)
		}
	}

	return result, nil
}

func DirLs(dir string) ([]interface{}, error) {
	result := []interface{}{}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return result, err
	}
	for _, f := range files {
		result = append(result, f)
	}
	return result, nil
}

func LoadResource(location string, network bool, urls []string) ([]byte, error) {
	var bytes []byte
	err := ErrNotFound

	if strings.HasPrefix(location, "http:/") || strings.HasPrefix(location, "https:/") {
		if !network {
			return nil, ErrNoNetwork
		}
		resp, err := http.Get(location)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("non-200 http response: %d", resp.StatusCode)
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	} else if strings.HasPrefix(location, "/") {
		return ioutil.ReadFile(location)
	} else if len(location) > 0 {
		for _, url := range urls {
			ymlUrl := fmt.Sprintf("%s/%s/%s.yml", url, location[0:1], location)
			bytes, err = LoadResource(ymlUrl, network, []string{})
			if err == nil {
				log.Debugf("Loaded %s from %s", location, ymlUrl)
				return bytes, nil
			}
		}
	}

	return nil, err
}

func Map2KVPairs(m map[string]string) []string {
	r := make([]string, 0, len(m))
	for k, v := range m {
		r = append(r, k+"="+v)
	}
	return r
}

func KVPairs2Map(kvs []string) map[string]string {
	r := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		s := strings.SplitN(kv, "=", 2)
		r[s[0]] = s[1]
	}
	return r
}

func TrimSplitN(str, sep string, count int) []string {
	result := []string{}
	for _, part := range strings.SplitN(strings.TrimSpace(str), sep, count) {
		result = append(result, strings.TrimSpace(part))
	}

	return result
}

func TrimSplit(str, sep string) []string {
	return TrimSplitN(str, sep, -1)
}
