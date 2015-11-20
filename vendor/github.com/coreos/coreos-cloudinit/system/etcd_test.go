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

func TestEtcdUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Etcd
		units  []Unit
	}{
		{
			config.Etcd{},
			[]Unit{{config.Unit{
				Name:    "etcd.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{Name: "20-cloudinit.conf"}},
			}}},
		},
		{
			config.Etcd{
				Discovery:    "http://disco.example.com/foobar",
				PeerBindAddr: "127.0.0.1:7002",
			},
			[]Unit{{config.Unit{
				Name:    "etcd.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{
					Name: "20-cloudinit.conf",
					Content: `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`,
				}},
			}}},
		},
		{
			config.Etcd{
				Name:         "node001",
				Discovery:    "http://disco.example.com/foobar",
				PeerBindAddr: "127.0.0.1:7002",
			},
			[]Unit{{config.Unit{
				Name:    "etcd.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{
					Name: "20-cloudinit.conf",
					Content: `[Service]
Environment="ETCD_DISCOVERY=http://disco.example.com/foobar"
Environment="ETCD_NAME=node001"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`,
				}},
			}}},
		},
	} {
		units := Etcd{tt.config}.Units()
		if !reflect.DeepEqual(tt.units, units) {
			t.Errorf("bad units (%+v): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}
