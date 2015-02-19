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
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
)

func ProcessDigitalOceanNetconf(config []byte) ([]InterfaceGenerator, error) {
	log.Println("Processing DigitalOcean network config")
	if len(config) == 0 {
		return nil, nil
	}

	var cfg digitalocean.Metadata
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, err
	}

	log.Println("Parsing nameservers")
	nameservers, err := parseNameservers(cfg.DNS)
	if err != nil {
		return nil, err
	}
	log.Printf("Parsed %d nameservers\n", len(nameservers))

	log.Println("Parsing interfaces")
	generators, err := parseInterfaces(cfg.Interfaces, nameservers)
	if err != nil {
		return nil, err
	}
	log.Printf("Parsed %d network interfaces\n", len(generators))

	log.Println("Processed DigitalOcean network config")
	return generators, nil
}

func parseNameservers(cfg digitalocean.DNS) ([]net.IP, error) {
	nameservers := make([]net.IP, 0, len(cfg.Nameservers))
	for _, ns := range cfg.Nameservers {
		if ip := net.ParseIP(ns); ip == nil {
			return nil, fmt.Errorf("could not parse %q as nameserver IP address", ns)
		} else {
			nameservers = append(nameservers, ip)
		}
	}
	return nameservers, nil
}

func parseInterfaces(cfg digitalocean.Interfaces, nameservers []net.IP) ([]InterfaceGenerator, error) {
	generators := make([]InterfaceGenerator, 0, len(cfg.Public)+len(cfg.Private))
	for _, iface := range cfg.Public {
		if generator, err := parseInterface(iface, nameservers, true); err == nil {
			generators = append(generators, &physicalInterface{*generator})
		} else {
			return nil, err
		}
	}
	for _, iface := range cfg.Private {
		if generator, err := parseInterface(iface, []net.IP{}, false); err == nil {
			generators = append(generators, &physicalInterface{*generator})
		} else {
			return nil, err
		}
	}
	return generators, nil
}

func parseInterface(iface digitalocean.Interface, nameservers []net.IP, useRoute bool) (*logicalInterface, error) {
	routes := make([]route, 0)
	addresses := make([]net.IPNet, 0)
	if iface.IPv4 != nil {
		var ip, mask, gateway net.IP
		if ip = net.ParseIP(iface.IPv4.IPAddress); ip == nil {
			return nil, fmt.Errorf("could not parse %q as IPv4 address", iface.IPv4.IPAddress)
		}
		if mask = net.ParseIP(iface.IPv4.Netmask); mask == nil {
			return nil, fmt.Errorf("could not parse %q as IPv4 mask", iface.IPv4.Netmask)
		}
		addresses = append(addresses, net.IPNet{
			IP:   ip,
			Mask: net.IPMask(mask),
		})

		if useRoute {
			if gateway = net.ParseIP(iface.IPv4.Gateway); gateway == nil {
				return nil, fmt.Errorf("could not parse %q as IPv4 gateway", iface.IPv4.Gateway)
			}
			routes = append(routes, route{
				destination: net.IPNet{
					IP:   net.IPv4zero,
					Mask: net.IPMask(net.IPv4zero),
				},
				gateway: gateway,
			})
		}
	}
	if iface.IPv6 != nil {
		var ip, gateway net.IP
		if ip = net.ParseIP(iface.IPv6.IPAddress); ip == nil {
			return nil, fmt.Errorf("could not parse %q as IPv6 address", iface.IPv6.IPAddress)
		}
		addresses = append(addresses, net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(iface.IPv6.Cidr, net.IPv6len*8),
		})

		if useRoute {
			if gateway = net.ParseIP(iface.IPv6.Gateway); gateway == nil {
				return nil, fmt.Errorf("could not parse %q as IPv6 gateway", iface.IPv6.Gateway)
			}
			routes = append(routes, route{
				destination: net.IPNet{
					IP:   net.IPv6zero,
					Mask: net.IPMask(net.IPv6zero),
				},
				gateway: gateway,
			})
		}
	}

	hwaddr, err := net.ParseMAC(iface.MAC)
	if err != nil {
		return nil, err
	}

	if nameservers == nil {
		nameservers = []net.IP{}
	}

	return &logicalInterface{
		hwaddr: hwaddr,
		config: configMethodStatic{
			addresses:   addresses,
			nameservers: nameservers,
			routes:      routes,
		},
	}, nil
}
