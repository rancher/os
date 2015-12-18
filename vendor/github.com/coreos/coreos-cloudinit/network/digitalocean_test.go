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

	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
)

func TestParseNameservers(t *testing.T) {
	for _, tt := range []struct {
		dns digitalocean.DNS
		nss []net.IP
		err error
	}{
		{
			dns: digitalocean.DNS{},
			nss: []net.IP{},
		},
		{
			dns: digitalocean.DNS{Nameservers: []string{"1.2.3.4"}},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
		},
		{
			dns: digitalocean.DNS{Nameservers: []string{"bad"}},
			err: errors.New("could not parse \"bad\" as nameserver IP address"),
		},
	} {
		nss, err := parseNameservers(tt.dns)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.dns, tt.err, err)
		}
		if !reflect.DeepEqual(tt.nss, nss) {
			t.Fatalf("bad nameservers (%+v): want %#v, got %#v", tt.dns, tt.nss, nss)
		}
	}
}

func mkInvalidMAC() error {
	if isGo15 {
		return &net.AddrError{Err: "invalid MAC address", Addr: "bad"}
	} else {
		return errors.New("invalid MAC address: bad")
	}
}

func TestParseInterface(t *testing.T) {
	for _, tt := range []struct {
		cfg      digitalocean.Interface
		nss      []net.IP
		useRoute bool
		iface    *logicalInterface
		err      error
	}{
		{
			cfg: digitalocean.Interface{
				MAC: "bad",
			},
			err: mkInvalidMAC(),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
			},
			useRoute: true,
			nss:      []net.IP{net.ParseIP("1.2.3.4")},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses:   []net.IPNet{},
					nameservers: []net.IP{net.ParseIP("1.2.3.4")},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "bad",
					Netmask:   "255.255.0.0",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv4 address"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "bad",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv4 mask"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "ignoreme",
				},
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses: []net.IPNet{net.IPNet{
						IP:   net.ParseIP("1.2.3.4"),
						Mask: net.IPMask(net.ParseIP("255.255.0.0")),
					}},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "bad",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			err:      errors.New("could not parse \"bad\" as IPv4 gateway"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "5.6.7.8",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses: []net.IPNet{net.IPNet{
						IP:   net.ParseIP("1.2.3.4"),
						Mask: net.IPMask(net.ParseIP("255.255.0.0")),
					}},
					nameservers: []net.IP{},
					routes: []route{route{
						net.IPNet{IP: net.IPv4zero, Mask: net.IPMask(net.IPv4zero)},
						net.ParseIP("5.6.7.8"),
					}},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "bad",
					Cidr:      16,
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as IPv6 address"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "ignoreme",
				},
			},
			nss: []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses: []net.IPNet{net.IPNet{
						IP:   net.ParseIP("fe00::"),
						Mask: net.IPMask(net.ParseIP("ffff::")),
					}},
					nameservers: []net.IP{},
					routes:      []route{},
				},
			},
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "bad",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			err:      errors.New("could not parse \"bad\" as IPv6 gateway"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv6: &digitalocean.Address{
					IPAddress: "fe00::",
					Cidr:      16,
					Gateway:   "fe00:1234::",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses: []net.IPNet{net.IPNet{
						IP:   net.ParseIP("fe00::"),
						Mask: net.IPMask(net.ParseIP("ffff::")),
					}},
					nameservers: []net.IP{},
					routes: []route{route{
						net.IPNet{IP: net.IPv6zero, Mask: net.IPMask(net.IPv6zero)},
						net.ParseIP("fe00:1234::"),
					}},
				},
			},
		},

		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				AnchorIPv4: &digitalocean.Address{
					IPAddress: "bad",
					Netmask:   "255.255.0.0",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as anchor IPv4 address"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				AnchorIPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "bad",
				},
			},
			nss: []net.IP{},
			err: errors.New("could not parse \"bad\" as anchor IPv4 mask"),
		},
		{
			cfg: digitalocean.Interface{
				MAC: "01:23:45:67:89:AB",
				IPv4: &digitalocean.Address{
					IPAddress: "1.2.3.4",
					Netmask:   "255.255.0.0",
					Gateway:   "5.6.7.8",
				},
				AnchorIPv4: &digitalocean.Address{
					IPAddress: "7.8.9.10",
					Netmask:   "255.255.0.0",
				},
			},
			useRoute: true,
			nss:      []net.IP{},
			iface: &logicalInterface{
				hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
				config: configMethodStatic{
					addresses: []net.IPNet{
						{
							IP:   net.ParseIP("1.2.3.4"),
							Mask: net.IPMask(net.ParseIP("255.255.0.0")),
						},
						{
							IP:   net.ParseIP("7.8.9.10"),
							Mask: net.IPMask(net.ParseIP("255.255.0.0")),
						},
					},
					nameservers: []net.IP{},
					routes: []route{
						{
							destination: net.IPNet{IP: net.IPv4zero, Mask: net.IPMask(net.IPv4zero)},
							gateway:     net.ParseIP("5.6.7.8"),
						},
						{
							destination: net.IPNet{IP: net.IPv4zero, Mask: net.IPMask(net.IPv4zero)},
						},
					},
				},
			},
		},
	} {
		iface, err := parseInterface(tt.cfg, tt.nss, tt.useRoute)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.iface, iface) {
			t.Fatalf("bad interface (%+v): want %#v, got %#v", tt.cfg, tt.iface, iface)
		}
	}
}

func TestParseInterfaces(t *testing.T) {
	for _, tt := range []struct {
		cfg    digitalocean.Interfaces
		nss    []net.IP
		ifaces []InterfaceGenerator
		err    error
	}{
		{
			ifaces: []InterfaceGenerator{},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{net.ParseIP("1.2.3.4")},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "01:23:45:67:89:AB"}},
			},
			nss: []net.IP{net.ParseIP("1.2.3.4")},
			ifaces: []InterfaceGenerator{
				&physicalInterface{logicalInterface{
					hwaddr: net.HardwareAddr([]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab}),
					config: configMethodStatic{
						addresses:   []net.IPNet{},
						nameservers: []net.IP{},
						routes:      []route{},
					},
				}},
			},
		},
		{
			cfg: digitalocean.Interfaces{
				Public: []digitalocean.Interface{{MAC: "bad"}},
			},
			err: mkInvalidMAC(),
		},
		{
			cfg: digitalocean.Interfaces{
				Private: []digitalocean.Interface{{MAC: "bad"}},
			},
			err: mkInvalidMAC(),
		},
	} {
		ifaces, err := parseInterfaces(tt.cfg, tt.nss)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%+v): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.ifaces, ifaces) {
			t.Fatalf("bad interfaces (%+v): want %#v, got %#v", tt.cfg, tt.ifaces, ifaces)
		}
	}
}

func TestProcessDigitalOceanNetconf(t *testing.T) {
	for _, tt := range []struct {
		cfg    digitalocean.Metadata
		ifaces []InterfaceGenerator
		err    error
	}{
		{
			cfg: digitalocean.Metadata{
				DNS: digitalocean.DNS{
					Nameservers: []string{"bad"},
				},
			},
			err: errors.New("could not parse \"bad\" as nameserver IP address"),
		},
		{
			cfg: digitalocean.Metadata{
				Interfaces: digitalocean.Interfaces{
					Public: []digitalocean.Interface{
						digitalocean.Interface{
							IPv4: &digitalocean.Address{
								IPAddress: "bad",
							},
						},
					},
				},
			},
			err: errors.New("could not parse \"bad\" as IPv4 address"),
		},
		{
			ifaces: []InterfaceGenerator{},
		},
	} {
		ifaces, err := ProcessDigitalOceanNetconf(tt.cfg)
		if !errorsEqual(tt.err, err) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.cfg, tt.err, err)
		}
		if !reflect.DeepEqual(tt.ifaces, ifaces) {
			t.Fatalf("bad interfaces (%q): want %#v, got %#v", tt.cfg, tt.ifaces, ifaces)
		}
	}
}

func errorsEqual(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if (a != nil && b == nil) || (a == nil && b != nil) {
		return false
	}
	return (a.Error() == b.Error())
}
