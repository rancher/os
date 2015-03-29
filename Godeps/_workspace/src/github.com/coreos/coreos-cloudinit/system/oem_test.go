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
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestOEMFile(t *testing.T) {
	for _, tt := range []struct {
		config config.OEM
		file   *File
	}{
		{
			config.OEM{},
			nil,
		},
		{
			config.OEM{
				ID:           "rackspace",
				Name:         "Rackspace Cloud Servers",
				VersionID:    "168.0.0",
				HomeURL:      "https://www.rackspace.com/cloud/servers/",
				BugReportURL: "https://github.com/coreos/coreos-overlay",
			},
			&File{config.File{
				Path:               "etc/oem-release",
				RawFilePermissions: "0644",
				Content: `ID=rackspace
VERSION_ID=168.0.0
NAME="Rackspace Cloud Servers"
HOME_URL="https://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
`,
			}},
		},
	} {
		file, err := OEM{tt.config}.File()
		if err != nil {
			t.Errorf("bad error (%q): want %v, got %q", tt.config, nil, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad file (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
