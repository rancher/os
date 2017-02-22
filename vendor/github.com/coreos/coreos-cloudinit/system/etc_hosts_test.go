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
	"os"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestEtcdHostsFile(t *testing.T) {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	for _, tt := range []struct {
		config config.EtcHosts
		file   *File
		err    error
	}{
		{
			"invalid",
			nil,
			fmt.Errorf("Invalid option to manage_etc_hosts"),
		},
		{
			"localhost",
			&File{config.File{
				Content:            fmt.Sprintf("127.0.0.1 %s\n", hostname),
				Path:               "etc/hosts",
				RawFilePermissions: "0644",
			}},
			nil,
		},
	} {
		file, err := EtcHosts{tt.config}.File()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
