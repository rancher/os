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

package vmware

import (
	"fmt"
	"net"
	"strings"

	"github.com/rancher/os/config/cloudinit/config"
	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/netconf"
)

type readConfigFunction func(key string) (string, error)
type urlDownloadFunction func(url string) ([]byte, error)

type VMWare struct {
	ovfFileName string
	readConfig  readConfigFunction
	urlDownload urlDownloadFunction
	lastError   error
}

func (v VMWare) Finish() error {
	return nil
}

func (v VMWare) String() string {
	return fmt.Sprintf("%s: %s (lastError: %v)", v.Type(), v.ovfFileName, v.lastError)
}

func (v VMWare) AvailabilityChanges() bool {
	return false
}

func (v VMWare) ConfigRoot() string {
	return "/"
}

func (v VMWare) read(keytmpl string, args ...interface{}) (string, error) {
	key := fmt.Sprintf(keytmpl, args...)
	return v.readConfig(key)
}

func (v VMWare) FetchMetadata() (metadata datasource.Metadata, err error) {
	metadata.NetworkConfig = netconf.NetworkConfig{}
	metadata.Hostname, _ = v.readConfig("hostname")

	//netconf := map[string]string{}
	//saveConfig := func(key string, args ...interface{}) string {
	//	key = fmt.Sprintf(key, args...)
	//	val, _ := v.readConfig(key)
	//	if val != "" {
	//		netconf[key] = val
	//	}
	//	return val
	//}

	for i := 0; ; i++ {
		val, _ := v.read("dns.server.%d", i)
		if val == "" {
			break
		}
		metadata.NetworkConfig.DNS.Nameservers = append(metadata.NetworkConfig.DNS.Nameservers, val)
	}
	dnsServers, _ := v.read("dns.servers")
	for _, val := range strings.Split(dnsServers, ",") {
		if val == "" {
			break
		}
		metadata.NetworkConfig.DNS.Nameservers = append(metadata.NetworkConfig.DNS.Nameservers, val)
	}

	for i := 0; ; i++ {
		//if domain := saveConfig("dns.domain.%d", i); domain == "" {
		val, _ := v.read("dns.domain.%d", i)
		if val == "" {
			break
		}
		metadata.NetworkConfig.DNS.Search = append(metadata.NetworkConfig.DNS.Search, val)
	}
	dnsDomains, _ := v.read("dns.domains")
	for _, val := range strings.Split(dnsDomains, ",") {
		if val == "" {
			break
		}
		metadata.NetworkConfig.DNS.Search = append(metadata.NetworkConfig.DNS.Search, val)
	}

	metadata.NetworkConfig.Interfaces = make(map[string]netconf.InterfaceConfig)
	found := true
	for i := 0; found; i++ {
		found = false

		ethName := fmt.Sprintf("eth%d", i)
		netDevice := netconf.InterfaceConfig{
			DHCP:      true,
			Match:     ethName,
			Addresses: []string{},
		}
		//found = (saveConfig("interface.%d.name", i) != "") || found
		if val, _ := v.read("interface.%d.name", i); val != "" {
			netDevice.Match = val
			found = true
		}
		//found = (saveConfig("interface.%d.mac", i) != "") || found
		if val, _ := v.read("interface.%d.mac", i); val != "" {
			netDevice.Match = "mac:" + val
			found = true
		}
		//found = (saveConfig("interface.%d.dhcp", i) != "") || found
		if val, _ := v.read("interface.%d.dhcp", i); val != "" {
			netDevice.DHCP = (strings.ToLower(val) != "no")
			found = true
		}

		role, _ := v.read("interface.%d.role", i)
		for a := 0; ; a++ {
			address, _ := v.read("interface.%d.ip.%d.address", i, a)
			if address == "" {
				break
			}
			netmask, _ := v.read("interface.%d.ip.%d.netmask", i, a)
			if netmask != "" {
				ones, _ := net.IPMask(net.ParseIP(netmask).To4()).Size()
				address = fmt.Sprintf("%s/%d", address, ones)
			}
			netDevice.Addresses = append(netDevice.Addresses, address)
			found = true
			netDevice.DHCP = false

			ip, _, err := net.ParseCIDR(address)
			if err != nil {
				log.Error(err)
				//return metadata, err
			}

			switch role {
			case "public":
				if ip.To4() != nil {
					metadata.PublicIPv4 = ip
				} else {
					metadata.PublicIPv6 = ip
				}
			case "private":
				if ip.To4() != nil {
					metadata.PrivateIPv4 = ip
				} else {
					metadata.PrivateIPv6 = ip
				}
			case "":
			default:
				//return metadata, fmt.Errorf("unrecognized role: %q", role)
				log.Error(err)
			}
		}

		for r := 0; ; r++ {
			gateway, _ := v.read("interface.%d.route.%d.gateway", i, r)
			// TODO: do we really not do anything but default routing?
			//destination, _ := v.read("interface.%d.route.%d.destination", i, r)
			destination := ""

			if gateway == "" && destination == "" {
				break
			} else {
				netDevice.Gateway = gateway
				found = true
			}
		}
		if found {
			metadata.NetworkConfig.Interfaces[ethName] = netDevice
		}
	}

	return
}

func (v VMWare) FetchUserdata() ([]byte, error) {
	encoding, err := v.readConfig("cloud-init.data.encoding")
	if err != nil {
		return nil, err
	}

	data, err := v.readConfig("cloud-init.config.data")
	if err != nil {
		return nil, err
	}

	// Try to fallback to url if no explicit data
	if data == "" {
		url, err := v.readConfig("cloud-init.config.url")
		if err != nil {
			return nil, err
		}

		if url != "" {
			rawData, err := v.urlDownload(url)
			if err != nil {
				return nil, err
			}
			data = string(rawData)
		}
	}

	if encoding != "" {
		return config.DecodeContent(data, encoding)
	}
	return []byte(data), nil
}

func (v VMWare) Type() string {
	return "VMWare"
}
