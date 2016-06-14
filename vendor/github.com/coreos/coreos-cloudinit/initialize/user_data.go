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

package initialize

import (
	"errors"
	"log"

	"github.com/coreos/coreos-cloudinit/config"
)

var (
	ErrIgnitionConfig = errors.New("not a config (found Ignition)")
)

func ParseUserData(contents string) (interface{}, error) {
	if len(contents) == 0 {
		return nil, nil
	}

	switch {
	case config.IsScript(contents):
		log.Printf("Parsing user-data as script")
		return config.NewScript(contents)
	case config.IsCloudConfig(contents):
		log.Printf("Parsing user-data as cloud-config")
		cc, err := config.NewCloudConfig(contents)
		if err != nil {
			return nil, err
		}

		if err := cc.Decode(); err != nil {
			return nil, err
		}

		return cc, nil
	case config.IsIgnitionConfig(contents):
		return nil, ErrIgnitionConfig
	default:
		return nil, errors.New("Unrecognized user-data format")
	}
}
