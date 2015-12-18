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
	"errors"
	"net"
	"reflect"
	"testing"
)

func mustParseMac(mac net.HardwareAddr, err error) net.HardwareAddr {
	if err != nil {
		panic(err)
	}
	return mac
}

func TestProcessVMwareNetconf(t *testing.T) {
	tests := []struct {
		config map[string]string

		interfaces []InterfaceGenerator
		err        error
	}{
		{},
		{
			config: map[string]string{
				"interface.0.dhcp": "yes",
			},
			interfaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					config: configMethodDHCP{},
				}},
			},
		},
		{
			config: map[string]string{
				"interface.0.mac":  "00:11:22:33:44:55",
				"interface.0.dhcp": "yes",
			},
			interfaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: mustParseMac(net.ParseMAC("00:11:22:33:44:55")),
					config: configMethodDHCP{hwaddress: mustParseMac(net.ParseMAC("00:11:22:33:44:55"))},
				}},
			},
		},
		{
			config: map[string]string{
				"interface.0.name": "eth0",
				"interface.0.dhcp": "yes",
			},
			interfaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					name:   "eth0",
					config: configMethodDHCP{},
				}},
			},
		},
		{
			config: map[string]string{
				"interface.0.mac":                 "00:11:22:33:44:55",
				"interface.0.ip.0.address":        "10.0.0.100/24",
				"interface.0.route.0.gateway":     "10.0.0.1",
				"interface.0.route.0.destination": "0.0.0.0/0",
			},
			interfaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: mustParseMac(net.ParseMAC("00:11:22:33:44:55")),
					config: configMethodStatic{
						hwaddress: mustParseMac(net.ParseMAC("00:11:22:33:44:55")),
						addresses: []net.IPNet{net.IPNet{IP: net.ParseIP("10.0.0.100"), Mask: net.CIDRMask(24, net.IPv4len*8)}},
						// I realize how upset you must be that I am shoving an IPMask into an IP. This is because net.IPv4zero is
						// actually a magic IPv6 address which ruins our equality check. What's that? Just use IP::Equal()? I'd rather
						// DeepEqual just handle that for me, but until Go gets operator overloading, we are stuck with this.
						routes: []route{route{
							destination: net.IPNet{IP: net.IP(net.CIDRMask(0, net.IPv4len*8)), Mask: net.CIDRMask(0, net.IPv4len*8)},
							gateway:     net.ParseIP("10.0.0.1")},
						},
					},
				}},
			},
		},
		{
			config: map[string]string{
				"dns.server.0":                    "1.2.3.4",
				"dns.server.1":                    "5.6.7.8",
				"interface.0.mac":                 "00:11:22:33:44:55",
				"interface.0.ip.0.address":        "10.0.0.100/24",
				"interface.0.ip.1.address":        "10.0.0.101/24",
				"interface.0.route.0.gateway":     "10.0.0.1",
				"interface.0.route.0.destination": "0.0.0.0/0",
				"interface.1.name":                "eth0",
				"interface.1.ip.0.address":        "10.0.1.100/24",
				"interface.1.route.0.gateway":     "10.0.1.1",
				"interface.1.route.0.destination": "0.0.0.0/0",
				"interface.2.dhcp":                "yes",
				"interface.2.mac":                 "00:11:22:33:44:77",
			},
			interfaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: mustParseMac(net.ParseMAC("00:11:22:33:44:55")),
					config: configMethodStatic{
						hwaddress: mustParseMac(net.ParseMAC("00:11:22:33:44:55")),
						addresses: []net.IPNet{
							net.IPNet{IP: net.ParseIP("10.0.0.100"), Mask: net.CIDRMask(24, net.IPv4len*8)},
							net.IPNet{IP: net.ParseIP("10.0.0.101"), Mask: net.CIDRMask(24, net.IPv4len*8)},
						},
						routes: []route{route{
							destination: net.IPNet{IP: net.IP(net.CIDRMask(0, net.IPv4len*8)), Mask: net.CIDRMask(0, net.IPv4len*8)},
							gateway:     net.ParseIP("10.0.0.1")},
						},
						nameservers: []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")},
					},
				}},
				&physicalInterface{logicalInterface{
					name: "eth0",
					config: configMethodStatic{
						addresses: []net.IPNet{net.IPNet{IP: net.ParseIP("10.0.1.100"), Mask: net.CIDRMask(24, net.IPv4len*8)}},
						routes: []route{route{
							destination: net.IPNet{IP: net.IP(net.CIDRMask(0, net.IPv4len*8)), Mask: net.CIDRMask(0, net.IPv4len*8)},
							gateway:     net.ParseIP("10.0.1.1")},
						},
						nameservers: []net.IP{net.ParseIP("1.2.3.4"), net.ParseIP("5.6.7.8")},
					},
				}},
				&physicalInterface{logicalInterface{
					hwaddr: mustParseMac(net.ParseMAC("00:11:22:33:44:77")),
					config: configMethodDHCP{hwaddress: mustParseMac(net.ParseMAC("00:11:22:33:44:77"))},
				}},
			},
		},
		{
			config: map[string]string{"dns.server.0": "test dns"},
			err:    errors.New(`invalid nameserver: "test dns"`),
		},
	}

	for i, tt := range tests {
		interfaces, err := ProcessVMwareNetconf(tt.config)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, tt.err, err)
		}
		if !reflect.DeepEqual(tt.interfaces, interfaces) {
			t.Errorf("bad interfaces (#%d): want %#v, got %#v", i, tt.interfaces, interfaces)
			for _, iface := range tt.interfaces {
				t.Logf("  want: %#v", iface)
			}
			for _, iface := range interfaces {
				t.Logf("  got:  %#v", iface)
			}
		}
	}
}

func TestProcessAddressConfig(t *testing.T) {
	tests := []struct {
		config map[string]string
		prefix string

		addresses []net.IPNet
		err       error
	}{
		{},

		// static - ipv4
		{
			config: map[string]string{
				"ip.0.address": "10.0.0.100/24",
			},

			addresses: []net.IPNet{{IP: net.ParseIP("10.0.0.100"), Mask: net.CIDRMask(24, net.IPv4len*8)}},
		},
		{
			config: map[string]string{
				"this.is.a.prefix.ip.0.address": "10.0.0.100/24",
			},
			prefix: "this.is.a.prefix.",

			addresses: []net.IPNet{{IP: net.ParseIP("10.0.0.100"), Mask: net.CIDRMask(24, net.IPv4len*8)}},
		},
		{
			config: map[string]string{
				"ip.0.address": "10.0.0.100/24",
				"ip.1.address": "10.0.0.101/24",
				"ip.2.address": "10.0.0.102/24",
			},

			addresses: []net.IPNet{
				{IP: net.ParseIP("10.0.0.100"), Mask: net.CIDRMask(24, net.IPv4len*8)},
				{IP: net.ParseIP("10.0.0.101"), Mask: net.CIDRMask(24, net.IPv4len*8)},
				{IP: net.ParseIP("10.0.0.102"), Mask: net.CIDRMask(24, net.IPv4len*8)},
			},
		},

		// static - ipv6
		{
			config: map[string]string{
				"ip.0.address": "fe00::100/64",
			},

			addresses: []net.IPNet{{IP: net.ParseIP("fe00::100"), Mask: net.IPMask(net.CIDRMask(64, net.IPv6len*8))}},
		},
		{
			config: map[string]string{
				"ip.0.address": "fe00::100/64",
				"ip.1.address": "fe00::101/64",
				"ip.2.address": "fe00::102/64",
			},

			addresses: []net.IPNet{
				{IP: net.ParseIP("fe00::100"), Mask: net.CIDRMask(64, net.IPv6len*8)},
				{IP: net.ParseIP("fe00::101"), Mask: net.CIDRMask(64, net.IPv6len*8)},
				{IP: net.ParseIP("fe00::102"), Mask: net.CIDRMask(64, net.IPv6len*8)},
			},
		},

		// invalid
		{
			config: map[string]string{
				"ip.0.address": "test address",
			},

			err: errors.New(`invalid address: "test address"`),
		},
	}

	for i, tt := range tests {
		addresses, err := processAddressConfig(tt.config, tt.prefix)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, tt.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(tt.addresses, addresses) {
			t.Errorf("bad addresses (#%d): want %#v, got %#v", i, tt.addresses, addresses)
		}
	}
}

func TestProcessRouteConfig(t *testing.T) {
	tests := []struct {
		config map[string]string
		prefix string

		routes []route
		err    error
	}{
		{},

		{
			config: map[string]string{
				"route.0.gateway":     "10.0.0.1",
				"route.0.destination": "0.0.0.0/0",
			},

			routes: []route{{destination: net.IPNet{IP: net.IP(net.CIDRMask(0, net.IPv4len*8)), Mask: net.CIDRMask(0, net.IPv4len*8)}, gateway: net.ParseIP("10.0.0.1")}},
		},
		{
			config: map[string]string{
				"this.is.a.prefix.route.0.gateway":     "10.0.0.1",
				"this.is.a.prefix.route.0.destination": "0.0.0.0/0",
			},
			prefix: "this.is.a.prefix.",

			routes: []route{{destination: net.IPNet{IP: net.IP(net.CIDRMask(0, net.IPv4len*8)), Mask: net.CIDRMask(0, net.IPv4len*8)}, gateway: net.ParseIP("10.0.0.1")}},
		},
		{
			config: map[string]string{
				"route.0.gateway":     "fe00::1",
				"route.0.destination": "::/0",
			},

			routes: []route{{destination: net.IPNet{IP: net.IPv6zero, Mask: net.IPMask(net.IPv6zero)}, gateway: net.ParseIP("fe00::1")}},
		},

		// invalid
		{
			config: map[string]string{
				"route.0.gateway":     "test gateway",
				"route.0.destination": "0.0.0.0/0",
			},

			err: errors.New(`invalid gateway: "test gateway"`),
		},
		{
			config: map[string]string{
				"route.0.gateway":     "10.0.0.1",
				"route.0.destination": "test destination",
			},

			err: &net.ParseError{Type: "CIDR address", Text: "test destination"},
		},
	}

	for i, tt := range tests {
		routes, err := processRouteConfig(tt.config, tt.prefix)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, tt.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(tt.routes, routes) {
			t.Errorf("bad routes (#%d): want %#v, got %#v", i, tt.routes, routes)
		}
	}
}

func TestProcessDHCPConfig(t *testing.T) {
	tests := []struct {
		config map[string]string
		prefix string

		dhcp bool
		err  error
	}{
		{},

		// prefix
		{config: map[string]string{"this.is.a.prefix.mac": ""}, prefix: "this.is.a.prefix.", dhcp: false},
		{config: map[string]string{"this.is.a.prefix.dhcp": "yes"}, prefix: "this.is.a.prefix.", dhcp: true},

		// dhcp
		{config: map[string]string{"dhcp": "yes"}, dhcp: true},
		{config: map[string]string{"dhcp": "no"}, dhcp: false},

		// invalid
		{config: map[string]string{"dhcp": "blah"}, err: errors.New(`invalid DHCP option: "blah"`)},
	}

	for i, tt := range tests {
		dhcp, err := processDHCPConfig(tt.config, tt.prefix)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (#%d): want %v, got %v", i, tt.err, err)
		}
		if err != nil {
			continue
		}

		if tt.dhcp != dhcp {
			t.Errorf("bad dhcp (#%d): want %v, got %v", i, tt.dhcp, dhcp)
		}
	}
}
