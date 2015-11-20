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

package system

import (
	"fmt"
	"reflect"

	"github.com/coreos/coreos-cloudinit/config"
)

// serviceContents generates the contents for a drop-in unit given the config.
// The argument must be a struct from the 'config' package.
func serviceContents(e interface{}) string {
	vars := getEnvVars(e)
	if len(vars) == 0 {
		return ""
	}

	out := "[Service]\n"
	for _, v := range vars {
		out += fmt.Sprintf("Environment=\"%s\"\n", v)
	}
	return out
}

func getEnvVars(e interface{}) []string {
	et := reflect.TypeOf(e)
	ev := reflect.ValueOf(e)

	vars := []string{}
	for i := 0; i < et.NumField(); i++ {
		if val := ev.Field(i).Interface(); !config.IsZero(val) {
			key := et.Field(i).Tag.Get("env")
			vars = append(vars, fmt.Sprintf("%s=%v", key, val))
		}
	}

	return vars
}
