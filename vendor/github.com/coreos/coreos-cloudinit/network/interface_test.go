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

package network

import (
	"net"
	"reflect"
	"testing"
)

func TestInterfaceGenerators(t *testing.T) {
	for _, tt := range []struct {
		name    string
		netdev  string
		link    string
		network string
		kind    string
		iface   InterfaceGenerator
	}{
		{
			name:    "",
			network: "[Match]\nMACAddress=00:01:02:03:04:05\n\n[Network]\n",
			kind:    "physical",
			iface: &physicalInterface{logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0, 1, 2, 3, 4, 5}),
			}},
		},
		{
			name:    "testname",
			network: "[Match]\nName=testname\n\n[Network]\nBond=testbond1\nVLAN=testvlan1\nVLAN=testvlan2\n",
			kind:    "physical",
			iface: &physicalInterface{logicalInterface{
				name: "testname",
				children: []networkInterface{
					&bondInterface{logicalInterface: logicalInterface{name: "testbond1"}},
					&vlanInterface{logicalInterface: logicalInterface{name: "testvlan1"}, id: 1},
					&vlanInterface{logicalInterface: logicalInterface{name: "testvlan2"}, id: 1},
				},
			}},
		},
		{
			name:    "testname",
			netdev:  "[NetDev]\nKind=bond\nName=testname\n\n[Bond]\n",
			network: "[Match]\nName=testname\n\n[Network]\nBond=testbond1\nVLAN=testvlan1\nVLAN=testvlan2\nDHCP=true\n",
			kind:    "bond",
			iface: &bondInterface{logicalInterface: logicalInterface{
				name:   "testname",
				config: configMethodDHCP{},
				children: []networkInterface{
					&bondInterface{logicalInterface: logicalInterface{name: "testbond1"}},
					&vlanInterface{logicalInterface: logicalInterface{name: "testvlan1"}, id: 1},
					&vlanInterface{logicalInterface: logicalInterface{name: "testvlan2"}, id: 1},
				},
			}},
		},
		{
			name:    "testname",
			netdev:  "[NetDev]\nKind=vlan\nName=testname\n\n[VLAN]\nId=1\n",
			network: "[Match]\nName=testname\n\n[Network]\n",
			kind:    "vlan",
			iface:   &vlanInterface{logicalInterface{name: "testname"}, 1, ""},
		},
		{
			name:    "testname",
			netdev:  "[NetDev]\nKind=vlan\nName=testname\nMACAddress=00:01:02:03:04:05\n\n[VLAN]\nId=1\n",
			network: "[Match]\nName=testname\n\n[Network]\n",
			kind:    "vlan",
			iface:   &vlanInterface{logicalInterface{name: "testname", config: configMethodStatic{hwaddress: net.HardwareAddr([]byte{0, 1, 2, 3, 4, 5})}}, 1, ""},
		},
		{
			name:    "testname",
			netdev:  "[NetDev]\nKind=vlan\nName=testname\nMACAddress=00:01:02:03:04:05\n\n[VLAN]\nId=1\n",
			network: "[Match]\nName=testname\n\n[Network]\nDHCP=true\n",
			kind:    "vlan",
			iface:   &vlanInterface{logicalInterface{name: "testname", config: configMethodDHCP{hwaddress: net.HardwareAddr([]byte{0, 1, 2, 3, 4, 5})}}, 1, ""},
		},
		{
			name:    "testname",
			netdev:  "[NetDev]\nKind=vlan\nName=testname\n\n[VLAN]\nId=0\n",
			network: "[Match]\nName=testname\n\n[Network]\nDNS=8.8.8.8\n\n[Address]\nAddress=192.168.1.100/24\n\n[Route]\nDestination=0.0.0.0/0\nGateway=1.2.3.4\n",
			kind:    "vlan",
			iface: &vlanInterface{logicalInterface: logicalInterface{
				name: "testname",
				config: configMethodStatic{
					addresses:   []net.IPNet{{IP: []byte{192, 168, 1, 100}, Mask: []byte{255, 255, 255, 0}}},
					nameservers: []net.IP{[]byte{8, 8, 8, 8}},
					routes:      []route{route{destination: net.IPNet{IP: []byte{0, 0, 0, 0}, Mask: []byte{0, 0, 0, 0}}, gateway: []byte{1, 2, 3, 4}}},
				},
			}},
		},
	} {
		if name := tt.iface.Name(); name != tt.name {
			t.Fatalf("bad name (%q): want %q, got %q", tt.iface, tt.name, name)
		}
		if netdev := tt.iface.Netdev(); netdev != tt.netdev {
			t.Fatalf("bad netdev (%q): want %q, got %q", tt.iface, tt.netdev, netdev)
		}
		if link := tt.iface.Link(); link != tt.link {
			t.Fatalf("bad link (%q): want %q, got %q", tt.iface, tt.link, link)
		}
		if network := tt.iface.Network(); network != tt.network {
			t.Fatalf("bad network (%q): want %q, got %q", tt.iface, tt.network, network)
		}
		if kind := tt.iface.Type(); kind != tt.kind {
			t.Fatalf("bad type (%q): want %q, got %q", tt.iface, tt.kind, kind)
		}
	}
}

func TestModprobeParams(t *testing.T) {
	for _, tt := range []struct {
		i InterfaceGenerator
		p string
	}{
		{
			i: &physicalInterface{},
			p: "",
		},
		{
			i: &vlanInterface{},
			p: "",
		},
		{
			i: &bondInterface{
				logicalInterface{},
				nil,
				map[string]string{
					"a": "1",
					"b": "2",
				},
			},
			p: "a=1 b=2",
		},
	} {
		if p := tt.i.ModprobeParams(); p != tt.p {
			t.Fatalf("bad params (%q): got %s, want %s", tt.i, p, tt.p)
		}
	}
}

func TestBuildInterfacesLo(t *testing.T) {
	stanzas := []*stanzaInterface{
		&stanzaInterface{
			name:         "lo",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodLoopback{},
			options:      map[string][]string{},
		},
	}
	interfaces := buildInterfaces(stanzas)
	if len(interfaces) != 0 {
		t.FailNow()
	}
}

func TestBuildInterfacesBlindBond(t *testing.T) {
	stanzas := []*stanzaInterface{
		{
			name:         "bond0",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"eth0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	bond0 := &bondInterface{
		logicalInterface{
			name:        "bond0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		[]string{"eth0"},
		map[string]string{},
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{bond0},
			configDepth: 1,
		},
	}
	expect := []InterfaceGenerator{bond0, eth0}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestBuildInterfacesBlindVLAN(t *testing.T) {
	stanzas := []*stanzaInterface{
		{
			name:         "vlan0",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"0"},
				"raw_device": []string{"eth0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	vlan0 := &vlanInterface{
		logicalInterface{
			name:        "vlan0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		0,
		"eth0",
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{vlan0},
			configDepth: 1,
		},
	}
	expect := []InterfaceGenerator{eth0, vlan0}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestBuildInterfaces(t *testing.T) {
	stanzas := []*stanzaInterface{
		&stanzaInterface{
			name:         "eth0",
			kind:         interfacePhysical,
			auto:         false,
			configMethod: configMethodManual{},
			options:      map[string][]string{},
		},
		&stanzaInterface{
			name:         "bond0",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"eth0"},
				"bond-mode":   []string{"4"},
				"bond-miimon": []string{"100"},
			},
		},
		&stanzaInterface{
			name:         "bond1",
			kind:         interfaceBond,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"bond-slaves": []string{"bond0"},
			},
		},
		&stanzaInterface{
			name:         "vlan0",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"0"},
				"raw_device": []string{"eth0"},
			},
		},
		&stanzaInterface{
			name:         "vlan1",
			kind:         interfaceVLAN,
			auto:         false,
			configMethod: configMethodManual{},
			options: map[string][]string{
				"id":         []string{"1"},
				"raw_device": []string{"bond0"},
			},
		},
	}
	interfaces := buildInterfaces(stanzas)
	vlan1 := &vlanInterface{
		logicalInterface{
			name:        "vlan1",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		1,
		"bond0",
	}
	vlan0 := &vlanInterface{
		logicalInterface{
			name:        "vlan0",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		0,
		"eth0",
	}
	bond1 := &bondInterface{
		logicalInterface{
			name:        "bond1",
			config:      configMethodManual{},
			children:    []networkInterface{},
			configDepth: 0,
		},
		[]string{"bond0"},
		map[string]string{},
	}
	bond0 := &bondInterface{
		logicalInterface{
			name:        "bond0",
			config:      configMethodManual{},
			children:    []networkInterface{bond1, vlan1},
			configDepth: 1,
		},
		[]string{"eth0"},
		map[string]string{
			"mode":   "4",
			"miimon": "100",
		},
	}
	eth0 := &physicalInterface{
		logicalInterface{
			name:        "eth0",
			config:      configMethodManual{},
			children:    []networkInterface{bond0, vlan0},
			configDepth: 2,
		},
	}
	expect := []InterfaceGenerator{bond0, bond1, eth0, vlan0, vlan1}
	if !reflect.DeepEqual(interfaces, expect) {
		t.FailNow()
	}
}

func TestFilename(t *testing.T) {
	for _, tt := range []struct {
		i logicalInterface
		f string
	}{
		{logicalInterface{name: "iface", configDepth: 0}, "00-iface"},
		{logicalInterface{name: "iface", configDepth: 9}, "09-iface"},
		{logicalInterface{name: "iface", configDepth: 10}, "0a-iface"},
		{logicalInterface{name: "iface", configDepth: 53}, "35-iface"},
		{logicalInterface{hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}), configDepth: 1}, "01-01:23:45:67:89:ab"},
		{logicalInterface{name: "iface", hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}), configDepth: 1}, "01-iface"},
	} {
		if tt.i.Filename() != tt.f {
			t.Fatalf("bad filename (%q): got %q, want %q", tt.i, tt.i.Filename(), tt.f)
		}
	}
}
