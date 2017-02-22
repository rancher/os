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
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestType(t *testing.T) {
	tests := []struct {
		name string

		typ string
	}{
		{},
		{"test.service", "service"},
		{"hello", ""},
		{"lots.of.dots", "dots"},
	}

	for _, tt := range tests {
		u := Unit{config.Unit{
			Name: tt.name,
		}}
		if typ := u.Type(); tt.typ != typ {
			t.Errorf("bad type (%+v): want %q, got %q", tt, tt.typ, typ)
		}
	}
}

func TestGroup(t *testing.T) {
	tests := []struct {
		name string

		group string
	}{
		{"test.service", "system"},
		{"test.link", "network"},
		{"test.network", "network"},
		{"test.netdev", "network"},
		{"test.conf", "system"},
	}

	for _, tt := range tests {
		u := Unit{config.Unit{
			Name: tt.name,
		}}
		if group := u.Group(); tt.group != group {
			t.Errorf("bad group (%+v): want %q, got %q", tt, tt.group, group)
		}
	}
}

func TestDestination(t *testing.T) {
	tests := []struct {
		root    string
		name    string
		runtime bool

		destination string
	}{
		{
			root:        "/some/dir",
			name:        "foobar.service",
			destination: "/some/dir/etc/systemd/system/foobar.service",
		},
		{
			root:        "/some/dir",
			name:        "foobar.service",
			runtime:     true,
			destination: "/some/dir/run/systemd/system/foobar.service",
		},
	}

	for _, tt := range tests {
		u := Unit{config.Unit{
			Name:    tt.name,
			Runtime: tt.runtime,
		}}
		if d := u.Destination(tt.root); tt.destination != d {
			t.Errorf("bad destination (%+v): want %q, got %q", tt, tt.destination, d)
		}
	}
}

func TestDropInDestination(t *testing.T) {
	tests := []struct {
		root       string
		unitName   string
		dropInName string
		runtime    bool

		destination string
	}{
		{
			root:        "/some/dir",
			unitName:    "foo.service",
			dropInName:  "bar.conf",
			destination: "/some/dir/etc/systemd/system/foo.service.d/bar.conf",
		},
		{
			root:        "/some/dir",
			unitName:    "foo.service",
			dropInName:  "bar.conf",
			runtime:     true,
			destination: "/some/dir/run/systemd/system/foo.service.d/bar.conf",
		},
	}

	for _, tt := range tests {
		u := Unit{config.Unit{
			Name:    tt.unitName,
			Runtime: tt.runtime,
			DropIns: []config.UnitDropIn{{
				Name: tt.dropInName,
			}},
		}}
		if d := u.DropInDestination(tt.root, u.DropIns[0]); tt.destination != d {
			t.Errorf("bad destination (%+v): want %q, got %q", tt, tt.destination, d)
		}
	}
}
