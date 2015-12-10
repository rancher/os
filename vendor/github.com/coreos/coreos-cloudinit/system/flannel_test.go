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

func TestFlannelEnvVars(t *testing.T) {
	for _, tt := range []struct {
		config   config.Flannel
		contents string
	}{
		{
			config.Flannel{},
			"",
		},
		{
			config.Flannel{
				EtcdEndpoints: "http://12.34.56.78:4001",
				EtcdPrefix:    "/coreos.com/network/tenant1",
			},
			`FLANNELD_ETCD_ENDPOINTS=http://12.34.56.78:4001
FLANNELD_ETCD_PREFIX=/coreos.com/network/tenant1`,
		},
	} {
		out := Flannel{tt.config}.envVars()
		if out != tt.contents {
			t.Errorf("bad contents (%+v): want %q, got %q", tt, tt.contents, out)
		}
	}
}

func TestFlannelFile(t *testing.T) {
	for _, tt := range []struct {
		config config.Flannel
		file   *File
	}{
		{
			config.Flannel{},
			nil,
		},
		{
			config.Flannel{
				EtcdEndpoints: "http://12.34.56.78:4001",
				EtcdPrefix:    "/coreos.com/network/tenant1",
			},
			&File{config.File{
				Path:               "run/flannel/options.env",
				RawFilePermissions: "0644",
				Content: `FLANNELD_ETCD_ENDPOINTS=http://12.34.56.78:4001
FLANNELD_ETCD_PREFIX=/coreos.com/network/tenant1`,
			}},
		},
	} {
		file, _ := Flannel{tt.config}.File()
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
