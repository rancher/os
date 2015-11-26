package project

import (
	"fmt"
	"strings"

	"github.com/flynn/go-shlex"
)

// Stringorslice represents a string or an array of strings.
// TODO use docker/docker/pkg/stringutils.StrSlice once 1.9.x is released.
type Stringorslice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s Stringorslice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Stringorslice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts := make([]string, len(value))
		for k, v := range value {
			parts[k] = v.(string)
		}
		s.parts = parts
	case string:
		s.parts = []string{value}
	default:
		return fmt.Errorf("Failed to unmarshal Stringorslice: %#v", value)
	}
	return nil
}

// Len returns the number of parts of the Stringorslice.
func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

// Slice gets the parts of the StrSlice as a Slice of string.
func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

// NewStringorslice creates an Stringorslice based on the specified parts (as strings).
func NewStringorslice(parts ...string) Stringorslice {
	return Stringorslice{parts}
}

// Command represents a docker command, can be a string or an array of strings.
// FIXME why not use Stringorslice (type Command struct { Stringorslice }
type Command struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s Command) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *Command) UnmarshalYAML(tag string, value interface{}) error {
	var err error
	switch value := value.(type) {
	case []interface{}:
		parts := make([]string, len(value))
		for k, v := range value {
			parts[k] = v.(string)
		}
		s.parts = parts
	case string:
		s.parts, err = shlex.Split(value)
	default:
		return fmt.Errorf("Failed to unmarshal Command: %#v", value)
	}
	return err
}

// ToString returns the parts of the command as a string (joined by spaces).
func (s *Command) ToString() string {
	return strings.Join(s.parts, " ")
}

// Slice gets the parts of the Command as a Slice of string.
func (s *Command) Slice() []string {
	return s.parts
}

// NewCommand create a Command based on the specified parts (as strings).
func NewCommand(parts ...string) Command {
	return Command{parts}
}

// SliceorMap represents a slice or a map of strings.
type SliceorMap struct {
	parts map[string]string
}

// MarshalYAML implements the Marshaller interface.
func (s SliceorMap) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *SliceorMap) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case map[interface{}]interface{}:
		parts := map[string]string{}
		for k, v := range value {
			parts[k.(string)] = v.(string)
		}
		s.parts = parts
	case []interface{}:
		parts := map[string]string{}
		for _, str := range value {
			str := strings.TrimSpace(str.(string))
			keyValueSlice := strings.SplitN(str, "=", 2)

			key := keyValueSlice[0]
			val := ""
			if len(keyValueSlice) == 2 {
				val = keyValueSlice[1]
			}
			parts[key] = val
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal SliceorMap: %#v", value)
	}
	return nil
}

// MapParts get the parts of the SliceorMap as a Map of string.
func (s *SliceorMap) MapParts() map[string]string {
	if s == nil {
		return nil
	}
	return s.parts
}

// NewSliceorMap creates a new SliceorMap based on the specified parts (as map of string).
func NewSliceorMap(parts map[string]string) SliceorMap {
	return SliceorMap{parts}
}

// MaporEqualSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key=value' string.
type MaporEqualSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporEqualSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporEqualSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts := make([]string, len(value))
		for k, v := range value {
			parts[k] = v.(string)
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts := make([]string, 0, len(value))
		for k, v := range value {
			parts = append(parts, strings.Join([]string{k.(string), v.(string)}, "="))
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporEqualSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporEqualSlice as a Slice of string.
func (s *MaporEqualSlice) Slice() []string {
	return s.parts
}

// NewMaporEqualSlice creates a new MaporEqualSlice based on the specified parts.
func NewMaporEqualSlice(parts []string) MaporEqualSlice {
	return MaporEqualSlice{parts}
}

// MaporColonSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key:value' string.
type MaporColonSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporColonSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporColonSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts := make([]string, len(value))
		for k, v := range value {
			parts[k] = v.(string)
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts := make([]string, 0, len(value))
		for k, v := range value {
			parts = append(parts, strings.Join([]string{k.(string), v.(string)}, ":"))
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporColonSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporColonSlice as a Slice of string.
func (s *MaporColonSlice) Slice() []string {
	return s.parts
}

// NewMaporColonSlice creates a new MaporColonSlice based on the specified parts.
func NewMaporColonSlice(parts []string) MaporColonSlice {
	return MaporColonSlice{parts}
}

// MaporSpaceSlice represents a slice of strings that gets unmarshal from a
// YAML map into 'key value' string.
type MaporSpaceSlice struct {
	parts []string
}

// MarshalYAML implements the Marshaller interface.
func (s MaporSpaceSlice) MarshalYAML() (tag string, value interface{}, err error) {
	return "", s.parts, nil
}

// UnmarshalYAML implements the Unmarshaller interface.
func (s *MaporSpaceSlice) UnmarshalYAML(tag string, value interface{}) error {
	switch value := value.(type) {
	case []interface{}:
		parts := make([]string, len(value))
		for k, v := range value {
			parts[k] = v.(string)
		}
		s.parts = parts
	case map[interface{}]interface{}:
		parts := make([]string, 0, len(value))
		for k, v := range value {
			parts = append(parts, strings.Join([]string{k.(string), v.(string)}, " "))
		}
		s.parts = parts
	default:
		return fmt.Errorf("Failed to unmarshal MaporSpaceSlice: %#v", value)
	}
	return nil
}

// Slice gets the parts of the MaporSpaceSlice as a Slice of string.
func (s *MaporSpaceSlice) Slice() []string {
	return s.parts
}

// NewMaporSpaceSlice creates a new MaporSpaceSlice based on the specified parts.
func NewMaporSpaceSlice(parts []string) MaporSpaceSlice {
	return MaporSpaceSlice{parts}
}
