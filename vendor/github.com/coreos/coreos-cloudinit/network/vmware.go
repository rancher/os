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
	"fmt"
	"log"
	"net"
)

func ProcessVMwareNetconf(config map[string]string) ([]InterfaceGenerator, error) {
	log.Println("Processing VMware network config")

	log.Println("Parsing nameservers")
	var nameservers []net.IP
	for i := 0; ; i++ {
		if ipStr, ok := config[fmt.Sprintf("dns.server.%d", i)]; ok {
			if ip := net.ParseIP(ipStr); ip != nil {
				nameservers = append(nameservers, ip)
			} else {
				return nil, fmt.Errorf("invalid nameserver: %q", ipStr)
			}
		} else {
			break
		}
	}
	log.Printf("Parsed %d nameservers", len(nameservers))

	var interfaces []InterfaceGenerator
	for i := 0; ; i++ {
		var addresses []net.IPNet
		var routes []route
		var err error
		var dhcp bool
		iface := &physicalInterface{}

		log.Printf("Proccessing interface %d", i)

		log.Println("Processing DHCP")
		if dhcp, err = processDHCPConfig(config, fmt.Sprintf("interface.%d.", i)); err != nil {
			return nil, err
		}

		log.Println("Processing addresses")
		if as, err := processAddressConfig(config, fmt.Sprintf("interface.%d.", i)); err == nil {
			addresses = append(addresses, as...)
		} else {
			return nil, err
		}

		log.Println("Processing routes")
		if rs, err := processRouteConfig(config, fmt.Sprintf("interface.%d.", i)); err == nil {
			routes = append(routes, rs...)
		} else {
			return nil, err
		}

		if mac, ok := config[fmt.Sprintf("interface.%d.mac", i)]; ok {
			log.Printf("Parsing interface %d MAC address: %q", i, mac)
			if hwaddr, err := net.ParseMAC(mac); err == nil {
				iface.hwaddr = hwaddr
			} else {
				return nil, fmt.Errorf("error while parsing MAC address: %v", err)
			}
		}

		if name, ok := config[fmt.Sprintf("interface.%d.name", i)]; ok {
			log.Printf("Parsing interface %d name: %q", i, name)
			iface.name = name
		}

		if len(addresses) > 0 || len(routes) > 0 {
			iface.config = configMethodStatic{
				hwaddress:   iface.hwaddr,
				addresses:   addresses,
				nameservers: nameservers,
				routes:      routes,
			}
		} else if dhcp {
			iface.config = configMethodDHCP{
				hwaddress: iface.hwaddr,
			}
		} else {
			break
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

func processAddressConfig(config map[string]string, prefix string) (addresses []net.IPNet, err error) {
	for a := 0; ; a++ {
		prefix := fmt.Sprintf("%sip.%d.", prefix, a)

		addressStr, ok := config[prefix+"address"]
		if !ok {
			break
		}

		ip, network, err := net.ParseCIDR(addressStr)
		if err != nil {
			return nil, fmt.Errorf("invalid address: %q", addressStr)
		}
		addresses = append(addresses, net.IPNet{
			IP:   ip,
			Mask: network.Mask,
		})
	}

	return
}

func processRouteConfig(config map[string]string, prefix string) (routes []route, err error) {
	for r := 0; ; r++ {
		prefix := fmt.Sprintf("%sroute.%d.", prefix, r)

		gatewayStr, gok := config[prefix+"gateway"]
		destinationStr, dok := config[prefix+"destination"]
		if gok && !dok {
			return nil, fmt.Errorf("missing destination key")
		} else if !gok && dok {
			return nil, fmt.Errorf("missing gateway key")
		} else if !gok && !dok {
			break
		}

		gateway := net.ParseIP(gatewayStr)
		if gateway == nil {
			return nil, fmt.Errorf("invalid gateway: %q", gatewayStr)
		}

		_, destination, err := net.ParseCIDR(destinationStr)
		if err != nil {
			return nil, err
		}

		routes = append(routes, route{
			destination: *destination,
			gateway:     gateway,
		})
	}

	return
}

func processDHCPConfig(config map[string]string, prefix string) (dhcp bool, err error) {
	dhcpStr, ok := config[prefix+"dhcp"]
	if !ok {
		return false, nil
	}

	switch dhcpStr {
	case "yes":
		return true, nil
	case "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid DHCP option: %q", dhcpStr)
	}
}
