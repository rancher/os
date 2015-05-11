package project

import (
	"strings"

	"gopkg.in/yaml.v2"
)

type Stringorslice struct {
	parts []string
}

func (s *Stringorslice) MarshalYAML() (interface{}, error) {
	if s == nil {
		return nil, nil
	}
	bytes, err := yaml.Marshal(s.Slice())
	return string(bytes), err
}

func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}
	return err
}

func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewStringorslice(parts ...string) Stringorslice {
	return Stringorslice{parts}
}

type SliceorMap struct {
	parts map[string]string
}

func (s *SliceorMap) MarshalYAML() (interface{}, error) {
	if s == nil {
		return nil, nil
	}
	bytes, err := yaml.Marshal(s.MapParts())
	return string(bytes), err
}

func (s *SliceorMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	mapType := make(map[string]string)
	err := unmarshal(&mapType)
	if err == nil {
		s.parts = mapType
		return nil
	}

	var sliceType []string
	var keyValueSlice []string
	var key string
	var value string

	err = unmarshal(&sliceType)
	if err == nil {
		mapType = make(map[string]string)
		for _, slice := range sliceType {
			keyValueSlice = strings.Split(slice, "=") //split up key and value into []string
			key = keyValueSlice[0]
			value = keyValueSlice[1]
			mapType[key] = value
		}
		s.parts = mapType
		return nil
	}
	return err
}

func (s *SliceorMap) MapParts() map[string]string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewSliceorMap(parts map[string]string) SliceorMap {
	return SliceorMap{parts}
}

type Maporslice struct {
	parts []string
}

func (s *Maporslice) MarshalYAML() (interface{}, error) {
	if s == nil {
		return nil, nil
	}
	bytes, err := yaml.Marshal(s.Slice())
	return string(bytes), err
}

func (s *Maporslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, "="))
	}

	return nil
}

func (s *Maporslice) Slice() []string {
	return s.parts
}

func NewMaporslice(parts []string) Maporslice {
	return Maporslice{parts}
}
