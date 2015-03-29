// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/coreos/yaml"
)

// CloudConfig encapsulates the entire cloud-config configuration file and maps
// directly to YAML. Fields that cannot be set in the cloud-config (fields
// used for internal use) have the YAML tag '-' so that they aren't marshalled.
type CloudConfig struct {
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	CoreOS            CoreOS   `yaml:"coreos"`
	WriteFiles        []File   `yaml:"write_files"`
	Hostname          string   `yaml:"hostname"`
	Users             []User   `yaml:"users"`
	ManageEtcHosts    EtcHosts `yaml:"manage_etc_hosts"`
}

type CoreOS struct {
	Etcd      Etcd      `yaml:"etcd"`
	Flannel   Flannel   `yaml:"flannel"`
	Fleet     Fleet     `yaml:"fleet"`
	Locksmith Locksmith `yaml:"locksmith"`
	OEM       OEM       `yaml:"oem"`
	Update    Update    `yaml:"update"`
	Units     []Unit    `yaml:"units"`
}

func IsCloudConfig(userdata string) bool {
	header := strings.SplitN(userdata, "\n", 2)[0]

	// Explicitly trim the header so we can handle user-data from
	// non-unix operating systems. The rest of the file is parsed
	// by yaml, which correctly handles CRLF.
	header = strings.TrimSuffix(header, "\r")

	return (header == "#cloud-config")
}

// NewCloudConfig instantiates a new CloudConfig from the given contents (a
// string of YAML), returning any error encountered. It will ignore unknown
// fields but log encountering them.
func NewCloudConfig(contents string) (*CloudConfig, error) {
	yaml.UnmarshalMappingKeyTransform = func(nameIn string) (nameOut string) {
		return strings.Replace(nameIn, "-", "_", -1)
	}
	var cfg CloudConfig
	err := yaml.Unmarshal([]byte(contents), &cfg)
	return &cfg, err
}

func (cc CloudConfig) String() string {
	bytes, err := yaml.Marshal(cc)
	if err != nil {
		return ""
	}

	stringified := string(bytes)
	stringified = fmt.Sprintf("#cloud-config\n%s", stringified)

	return stringified
}

// IsZero returns whether or not the parameter is the zero value for its type.
// If the parameter is a struct, only the exported fields are considered.
func IsZero(c interface{}) bool {
	return isZero(reflect.ValueOf(c))
}

type ErrorValid struct {
	Value string
	Valid string
	Field string
}

func (e ErrorValid) Error() string {
	return fmt.Sprintf("invalid value %q for option %q (valid options: %q)", e.Value, e.Field, e.Valid)
}

// AssertStructValid checks the fields in the structure and makes sure that
// they contain valid values as specified by the 'valid' flag. Empty fields are
// implicitly valid.
func AssertStructValid(c interface{}) error {
	ct := reflect.TypeOf(c)
	cv := reflect.ValueOf(c)
	for i := 0; i < ct.NumField(); i++ {
		ft := ct.Field(i)
		if !isFieldExported(ft) {
			continue
		}

		if err := AssertValid(cv.Field(i), ft.Tag.Get("valid")); err != nil {
			err.Field = ft.Name
			return err
		}
	}
	return nil
}

// AssertValid checks to make sure that the given value is in the list of
// valid values. Zero values are implicitly valid.
func AssertValid(value reflect.Value, valid string) *ErrorValid {
	if valid == "" || isZero(value) {
		return nil
	}

	vs := fmt.Sprintf("%v", value.Interface())
	if m, _ := regexp.MatchString(valid, vs); m {
		return nil
	}

	return &ErrorValid{
		Value: vs,
		Valid: valid,
	}
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Struct:
		vt := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if isFieldExported(vt.Field(i)) && !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

func isFieldExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}
