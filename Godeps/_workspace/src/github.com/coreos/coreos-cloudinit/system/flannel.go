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
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

// flannel is a top-level structure which embeds its underlying configuration,
// config.Flannel, and provides the system-specific Unit().
type Flannel struct {
	config.Flannel
}

func (fl Flannel) envVars() string {
	return strings.Join(getEnvVars(fl.Flannel), "\n")
}

func (fl Flannel) File() (*File, error) {
	vars := fl.envVars()
	if vars == "" {
		return nil, nil
	}
	return &File{config.File{
		Path:               path.Join("run", "flannel", "options.env"),
		RawFilePermissions: "0644",
		Content:            vars,
	}}, nil
}
