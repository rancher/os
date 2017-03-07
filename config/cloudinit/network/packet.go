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

	"github.com/rancher/os/netconf"
)

func ProcessPacketNetconf(netdata netconf.NetworkConfig) ([]InterfaceGenerator, error) {
	var nameservers []net.IP
	for _, v := range netdata.DNS.Nameservers {
		nameservers = append(nameservers, net.ParseIP(v))
	}
	if len(nameservers) == 0 {
		nameservers = append(nameservers, net.ParseIP("8.8.8.8"), net.ParseIP("8.8.4.4"))
	}

	generators, err := parseNetwork(netdata, nameservers)
	if err != nil {
		return nil, err
	}

	return generators, nil
}

func parseNetwork(netdata netconf.NetworkConfig, nameservers []net.IP) ([]InterfaceGenerator, error) {
	var interfaces []InterfaceGenerator
	var addresses []net.IPNet
	var routes []route
	/*	for _, netblock := range netdata.Netblocks {
			addresses = append(addresses, net.IPNet{
				IP:   netblock.Address,
				Mask: net.IPMask(netblock.Netmask),
			})
			if netblock.Public == false {
				routes = append(routes, route{
					destination: net.IPNet{
						IP:   net.IPv4(10, 0, 0, 0),
						Mask: net.IPv4Mask(255, 0, 0, 0),
					},
					gateway: netblock.Gateway,
				})
			} else {
				if netblock.AddressFamily == 4 {
					routes = append(routes, route{
						destination: net.IPNet{
							IP:   net.IPv4zero,
							Mask: net.IPMask(net.IPv4zero),
						},
						gateway: netblock.Gateway,
					})
				} else {
					routes = append(routes, route{
						destination: net.IPNet{
							IP:   net.IPv6zero,
							Mask: net.IPMask(net.IPv6zero),
						},
						gateway: netblock.Gateway,
					})
				}
			}
		}
	*/

	bond := bondInterface{
		logicalInterface: logicalInterface{
			name: "bond0",
			config: configMethodStatic{
				addresses:   addresses,
				nameservers: nameservers,
				routes:      routes,
			},
		},
		options: map[string]string{
			"Mode":             "802.3ad",
			"LACPTransmitRate": "fast",
			"MIIMonitorSec":    ".2",
			"UpDelaySec":       ".2",
			"DownDelaySec":     ".2",
		},
	}

	//bond.hwaddr, _ = net.ParseMAC(netdata.Interfaces[0].Mac)

	index := 0
	for name := range netdata.Interfaces {
		bond.slaves = append(bond.slaves, name)

		interfaces = append(interfaces, &physicalInterface{
			logicalInterface: logicalInterface{
				name: name,
				config: configMethodStatic{
					nameservers: nameservers,
				},
				children:    []networkInterface{&bond},
				configDepth: index,
			},
		})
		index = index + 1
	}

	interfaces = append(interfaces, &bond)

	return interfaces, nil
}
