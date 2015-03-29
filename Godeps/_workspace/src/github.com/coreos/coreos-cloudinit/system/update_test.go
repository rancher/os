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
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func testReadConfig(config string) func() (io.Reader, error) {
	return func() (io.Reader, error) {
		return strings.NewReader(config), nil
	}
}

func TestUpdateUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Update
		units  []Unit
		err    error
	}{
		{
			config: config.Update{},
		},
		{
			config: config.Update{Group: "master", Server: "http://foo.com"},
			units: []Unit{{config.Unit{
				Name:    "update-engine.service",
				Command: "restart",
			}}},
		},
		{
			config: config.Update{RebootStrategy: "best-effort"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "reboot"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "restart",
				Runtime: true,
			}}},
		},
		{
			config: config.Update{RebootStrategy: "off"},
			units: []Unit{{config.Unit{
				Name:    "locksmithd.service",
				Command: "stop",
				Runtime: true,
				Mask:    true,
			}}},
		},
	} {
		units := Update{Update: tt.config, ReadConfig: testReadConfig("")}.Units()
		if !reflect.DeepEqual(tt.units, units) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}

func TestUpdateFile(t *testing.T) {
	for _, tt := range []struct {
		config config.Update
		orig   string
		file   *File
		err    error
	}{
		{
			config: config.Update{},
		},
		{
			config: config.Update{RebootStrategy: "wizzlewazzle"},
			err:    &config.ErrorValid{Value: "wizzlewazzle", Field: "RebootStrategy", Valid: "^(best-effort|etcd-lock|reboot|off)$"},
		},
		{
			config: config.Update{Group: "master", Server: "http://foo.com"},
			file: &File{config.File{
				Content:            "GROUP=master\nSERVER=http://foo.com\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
		{
			config: config.Update{RebootStrategy: "best-effort"},
			file: &File{config.File{
				Content:            "REBOOT_STRATEGY=best-effort\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			file: &File{config.File{
				Content:            "REBOOT_STRATEGY=etcd-lock\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
		{
			config: config.Update{RebootStrategy: "reboot"},
			file: &File{config.File{
				Content:            "REBOOT_STRATEGY=reboot\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
		{
			config: config.Update{RebootStrategy: "off"},
			file: &File{config.File{
				Content:            "REBOOT_STRATEGY=off\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
		{
			config: config.Update{RebootStrategy: "etcd-lock"},
			orig:   "SERVER=https://example.com\nGROUP=thegroupc\nREBOOT_STRATEGY=awesome",
			file: &File{config.File{
				Content:            "SERVER=https://example.com\nGROUP=thegroupc\nREBOOT_STRATEGY=etcd-lock\n",
				Path:               "etc/coreos/update.conf",
				RawFilePermissions: "0644",
			}},
		},
	} {
		file, err := Update{Update: tt.config, ReadConfig: testReadConfig(tt.orig)}.File()
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%q): want %q, got %q", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.file, file) {
			t.Errorf("bad units (%q): want %#v, got %#v", tt.config, tt.file, file)
		}
	}
}
