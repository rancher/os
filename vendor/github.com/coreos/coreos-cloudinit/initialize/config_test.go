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
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/system"
)

type TestUnitManager struct {
	placed   []string
	enabled  []string
	masked   []string
	unmasked []string
	commands []UnitAction
	reload   bool
}

type UnitAction struct {
	unit    string
	command string
}

func (tum *TestUnitManager) PlaceUnit(u system.Unit) error {
	tum.placed = append(tum.placed, u.Name)
	return nil
}
func (tum *TestUnitManager) PlaceUnitDropIn(u system.Unit, d config.UnitDropIn) error {
	tum.placed = append(tum.placed, u.Name+".d/"+d.Name)
	return nil
}
func (tum *TestUnitManager) EnableUnitFile(u system.Unit) error {
	tum.enabled = append(tum.enabled, u.Name)
	return nil
}
func (tum *TestUnitManager) RunUnitCommand(u system.Unit, c string) (string, error) {
	tum.commands = append(tum.commands, UnitAction{u.Name, c})
	return "", nil
}
func (tum *TestUnitManager) DaemonReload() error {
	tum.reload = true
	return nil
}
func (tum *TestUnitManager) MaskUnit(u system.Unit) error {
	tum.masked = append(tum.masked, u.Name)
	return nil
}
func (tum *TestUnitManager) UnmaskUnit(u system.Unit) error {
	tum.unmasked = append(tum.unmasked, u.Name)
	return nil
}

type mockInterface struct {
	name           string
	filename       string
	netdev         string
	link           string
	network        string
	kind           string
	modprobeParams string
}

func (i mockInterface) Name() string {
	return i.name
}

func (i mockInterface) Filename() string {
	return i.filename
}

func (i mockInterface) Netdev() string {
	return i.netdev
}

func (i mockInterface) Link() string {
	return i.link
}

func (i mockInterface) Network() string {
	return i.network
}

func (i mockInterface) Type() string {
	return i.kind
}

func (i mockInterface) ModprobeParams() string {
	return i.modprobeParams
}

func TestCreateNetworkingUnits(t *testing.T) {
	for _, tt := range []struct {
		interfaces []network.InterfaceGenerator
		expect     []system.Unit
	}{
		{nil, nil},
		{
			[]network.InterfaceGenerator{
				network.InterfaceGenerator(mockInterface{filename: "test"}),
			},
			nil,
		},
		{
			[]network.InterfaceGenerator{
				network.InterfaceGenerator(mockInterface{filename: "test1", netdev: "test netdev"}),
				network.InterfaceGenerator(mockInterface{filename: "test2", link: "test link"}),
				network.InterfaceGenerator(mockInterface{filename: "test3", network: "test network"}),
			},
			[]system.Unit{
				system.Unit{Unit: config.Unit{Name: "test1.netdev", Runtime: true, Content: "test netdev"}},
				system.Unit{Unit: config.Unit{Name: "test2.link", Runtime: true, Content: "test link"}},
				system.Unit{Unit: config.Unit{Name: "test3.network", Runtime: true, Content: "test network"}},
			},
		},
		{
			[]network.InterfaceGenerator{
				network.InterfaceGenerator(mockInterface{filename: "test", netdev: "test netdev", link: "test link", network: "test network"}),
			},
			[]system.Unit{
				system.Unit{Unit: config.Unit{Name: "test.netdev", Runtime: true, Content: "test netdev"}},
				system.Unit{Unit: config.Unit{Name: "test.link", Runtime: true, Content: "test link"}},
				system.Unit{Unit: config.Unit{Name: "test.network", Runtime: true, Content: "test network"}},
			},
		},
	} {
		units := createNetworkingUnits(tt.interfaces)
		if !reflect.DeepEqual(tt.expect, units) {
			t.Errorf("bad units (%+v): want %#v, got %#v", tt.interfaces, tt.expect, units)
		}
	}
}

func TestProcessUnits(t *testing.T) {
	tests := []struct {
		units []system.Unit

		result TestUnitManager
	}{
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name: "foo",
					Mask: true,
				}},
			},
			result: TestUnitManager{
				masked: []string{"foo"},
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name:    "baz.service",
					Content: "[Service]\nExecStart=/bin/baz",
					Command: "start",
				}},
				system.Unit{Unit: config.Unit{
					Name:    "foo.network",
					Content: "[Network]\nFoo=true",
				}},
				system.Unit{Unit: config.Unit{
					Name:    "bar.network",
					Content: "[Network]\nBar=true",
				}},
			},
			result: TestUnitManager{
				placed: []string{"baz.service", "foo.network", "bar.network"},
				commands: []UnitAction{
					UnitAction{"systemd-networkd.service", "restart"},
					UnitAction{"baz.service", "start"},
				},
				reload: true,
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name:    "baz.service",
					Content: "[Service]\nExecStart=/bin/true",
				}},
			},
			result: TestUnitManager{
				placed: []string{"baz.service"},
				reload: true,
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name:    "locksmithd.service",
					Runtime: true,
				}},
			},
			result: TestUnitManager{
				unmasked: []string{"locksmithd.service"},
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name:   "woof",
					Enable: true,
				}},
			},
			result: TestUnitManager{
				enabled: []string{"woof"},
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name:    "hi.service",
					Runtime: true,
					Content: "[Service]\nExecStart=/bin/echo hi",
					DropIns: []config.UnitDropIn{
						{
							Name:    "lo.conf",
							Content: "[Service]\nExecStart=/bin/echo lo",
						},
						{
							Name:    "bye.conf",
							Content: "[Service]\nExecStart=/bin/echo bye",
						},
					},
				}},
			},
			result: TestUnitManager{
				placed:   []string{"hi.service", "hi.service.d/lo.conf", "hi.service.d/bye.conf"},
				unmasked: []string{"hi.service"},
				reload:   true,
			},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					DropIns: []config.UnitDropIn{
						{
							Name:    "lo.conf",
							Content: "[Service]\nExecStart=/bin/echo lo",
						},
					},
				}},
			},
			result: TestUnitManager{},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name: "hi.service",
					DropIns: []config.UnitDropIn{
						{
							Content: "[Service]\nExecStart=/bin/echo lo",
						},
					},
				}},
			},
			result: TestUnitManager{},
		},
		{
			units: []system.Unit{
				system.Unit{Unit: config.Unit{
					Name: "hi.service",
					DropIns: []config.UnitDropIn{
						{
							Name: "lo.conf",
						},
					},
				}},
			},
			result: TestUnitManager{},
		},
	}

	for _, tt := range tests {
		tum := &TestUnitManager{}
		if err := processUnits(tt.units, "", tum); err != nil {
			t.Errorf("bad error (%+v): want nil, got %s", tt.units, err)
		}
		if !reflect.DeepEqual(tt.result, *tum) {
			t.Errorf("bad result (%+v): want %+v, got %+v", tt.units, tt.result, tum)
		}
	}
}
