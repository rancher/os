package yaml

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/engine-api/types/strslice"
	"github.com/flynn/go-shlex"
)

// Stringorslice represents a string or an array of strings.
// Using engine-api Strslice and augment it with YAML marshalling stuff.
type Stringorslice strslice.StrSlice

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		*s = parts
	case string:
		*s = []string{value}
	default:
		return fmt.Errorf("Failed to unmarshal Stringorslice: %#v", value)
	}
	return nil
}

// Ulimits represents a list of Ulimit.
// It is, however, represented in yaml as keys (and thus map in Go)
type Ulimits struct {
	Elements []Ulimit
}

// MarshalYAML implements the Marshaller interface.
func (u Ulimits) MarshalYAML() (tag string, value interface{}, err error) {
	ulimitMap := make(map[string]Ulimit)
	for _, ulimit := range u.Elements {
		ulimitMap[ulimit.Name] = ulimit
	}
	return "", ulimitMap, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (u *Ulimits) UnmarshalYAML(tag string, value interface{}) error {
	ulimits := make(map[string]Ulimit)
	yamlUlimits := reflect.ValueOf(value)
	switch yamlUlimits.Kind() {
	case reflect.Map:
		for _, key := range yamlUlimits.MapKeys() {
			var name string
			var soft, hard int64
			mapValue := yamlUlimits.MapIndex(key).Elem()
			name = key.Elem().String()
			switch mapValue.Kind() {
			case reflect.Int64:
				soft = mapValue.Int()
				hard = mapValue.Int()
			case reflect.Map:
				if len(mapValue.MapKeys()) != 2 {
					return fmt.Errorf("Failed to unmarshal Ulimit: %#v", mapValue)
				}
				for _, subKey := range mapValue.MapKeys() {
					subValue := mapValue.MapIndex(subKey).Elem()
					switch subKey.Elem().String() {
					case "soft":
						soft = subValue.Int()
					case "hard":
						hard = subValue.Int()
					}
				}
			default:
				return fmt.Errorf("Failed to unmarshal Ulimit: %#v, %v", mapValue, mapValue.Kind())
			}
			ulimits[name] = Ulimit{
				Name: name,
				ulimitValues: ulimitValues{
					Soft: soft,
					Hard: hard,
				},
			}
		}
		keys := make([]string, 0, len(ulimits))
		for key := range ulimits {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			u.Elements = append(u.Elements, ulimits[key])
		}
	default:
		return fmt.Errorf("Failed to unmarshal Ulimit: %#v", value)
	}
	return nil
}

// Ulimit represents ulimit information.
type Ulimit struct {
	ulimitValues
	Name string
}

type ulimitValues struct {
	Soft int64 `yaml:"soft"`
	Hard int64 `yaml:"hard"`
}

// MarshalYAML implements the Marshaller interface.
func (u Ulimit) MarshalYAML() (tag string, value interface{}, err error) {
	if u.Soft == u.Hard {
		return "", u.Soft, nil
	}
	return "", u.ulimitValues, err
}

// NewUlimit creates a Ulimit based on the specified parts.
func NewUlimit(name string, soft int64, hard int64) Ulimit {
	return Ulimit{
		Name: name,
		ulimitValues: ulimitValues{
			Soft: soft,
			Hard: hard,
		},
	}
}

// Command represents a docker command, can be a string or an array of strings.
type Command strslice.StrSlice

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Command) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts, err := toStrings(value)
		if err != nil {
			return err
		}
		*s = parts
	case string:
		parts, err := shlex.Split(value)
		if err != nil {
			return err
		}
		*s = parts
	default:
		return fmt.Errorf("Failed to unmarshal Command: %#v", value)
	}
	return nil
}

// SliceorMap represents a slice or a map of strings.
type SliceorMap map[string]string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *SliceorMap) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case map[interface{}]interface{}:
		parts := map[string]string{}
		for k, v := range value {
			if sk, ok := k.(string); ok {
				if sv, ok := v.(string); ok {
					parts[sk] = sv
				} else {
					return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
				}
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
			}
		}
		*s = parts
	case []interface{}:
		parts := map[string]string{}
		for _, s := range value {
			if str, ok := s.(string); ok {
				str := strings.TrimSpace(str)
				keyValueSlice := strings.SplitN(str, "=", 2)

				key := keyValueSlice[0]
				val := ""
				if len(keyValueSlice) == 2 {
					val = keyValueSlice[1]
				}
				parts[key] = val
			} else {
				return fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", s, s)
			}
		}
		*s = parts
	default:
		return fmt.Errorf("Failed to unmarshal SliceorMap: %#v", value)
	}
	return nil
}

// MaporEqualSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key=value' string.
type MaporEqualSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporEqualSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, "=")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

// MaporColonSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key:value' string.
type MaporColonSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporColonSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, ":")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

// MaporSpaceSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key value' string.
type MaporSpaceSlice []string

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporSpaceSlice) UnmarshalYAML(tag string, value interface{}) error {
	parts, err := unmarshalToStringOrSepMapParts(value, " ")
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

func unmarshalToStringOrSepMapParts(value interface{}, key string) ([]string, error) {
	switch value := value.(type) {
	case []interface{}:
		return toStrings(value)
	case map[interface{}]interface{}:
		return toSepMapParts(value, key)
	default:
		return nil, fmt.Errorf("Failed to unmarshal Map or Slice: %#v", value)
	}
}

func toSepMapParts(value map[interface{}]interface{}, sep string) ([]string, error) {
	if len(value) == 0 {
		return nil, nil
	}
	parts := make([]string, 0, len(value))
	for k, v := range value {
		if sk, ok := k.(string); ok {
			if sv, ok := v.(string); ok {
				parts = append(parts, sk+sep+sv)
			} else if sv, ok := v.(int64); ok {
				parts = append(parts, sk+sep+strconv.FormatInt(sv, 10))
			} else {
				return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
			}
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", k, k)
		}
	}
	return parts, nil
}

func toStrings(s []interface{}) ([]string, error) {
	if len(s) == 0 {
		return nil, nil
	}
	r := make([]string, len(s))
	for k, v := range s {
		if sv, ok := v.(string); ok {
			r[k] = sv
		} else {
			return nil, fmt.Errorf("Cannot unmarshal '%v' of type %T into a string value", v, v)
		}
	}
	return r, nil
}
