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

func TestLocksmithUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Locksmith
		units  []Unit
	}{
		{
			config.Locksmith{},
			[]Unit{{config.Unit{
				Name:    "locksmithd.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{Name: "20-cloudinit.conf"}},
			}}},
		},
		{
			config.Locksmith{
				Endpoint: "12.34.56.78:4001",
			},
			[]Unit{{config.Unit{
				Name:    "locksmithd.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{
					Name: "20-cloudinit.conf",
					Content: `[Service]
Environment="LOCKSMITHD_ENDPOINT=12.34.56.78:4001"
`,
				}},
			}}},
		},
	} {
		units := Locksmith{tt.config}.Units()
		if !reflect.DeepEqual(units, tt.units) {
			t.Errorf("bad units (%+v): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}
