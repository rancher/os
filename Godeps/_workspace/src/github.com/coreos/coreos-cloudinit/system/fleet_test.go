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

func TestFleetUnits(t *testing.T) {
	for _, tt := range []struct {
		config config.Fleet
		units  []Unit
	}{
		{
			config.Fleet{},
			[]Unit{{config.Unit{
				Name:    "fleet.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{Name: "20-cloudinit.conf"}},
			}}},
		},
		{
			config.Fleet{
				PublicIP: "12.34.56.78",
			},
			[]Unit{{config.Unit{
				Name:    "fleet.service",
				Runtime: true,
				DropIns: []config.UnitDropIn{{
					Name: "20-cloudinit.conf",
					Content: `[Service]
Environment="FLEET_PUBLIC_IP=12.34.56.78"
`,
				}},
			}}},
		},
	} {
		units := Fleet{tt.config}.Units()
		if !reflect.DeepEqual(units, tt.units) {
			t.Errorf("bad units (%+v): want %#v, got %#v", tt.config, tt.units, units)
		}
	}
}
