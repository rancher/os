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

	"github.com/rancher/os/config/cloudinit/config"
	"github.com/rancher/os/config/cloudinit/datasource"
)

type readConfigFunction func(key string) (string, error)
type urlDownloadFunction func(url string) ([]byte, error)

type vmware struct {
	ovfFileName string
	readConfig  readConfigFunction
	urlDownload urlDownloadFunction
}

func (v vmware) AvailabilityChanges() bool {
	return false
}

func (v vmware) ConfigRoot() string {
	return "/"
}

func (v vmware) FetchMetadata() (metadata datasource.Metadata, err error) {
	metadata.Hostname, _ = v.readConfig("hostname")

	netconf := map[string]string{}
	saveConfig := func(key string, args ...interface{}) string {
		key = fmt.Sprintf(key, args...)
		val, _ := v.readConfig(key)
		if val != "" {
			netconf[key] = val
		}
		return val
	}

	for i := 0; ; i++ {
		if nameserver := saveConfig("dns.server.%d", i); nameserver == "" {
			break
		}
	}

	for i := 0; ; i++ {
		if domain := saveConfig("dns.domain.%d", i); domain == "" {
			break
		}
	}

	found := true
	for i := 0; found; i++ {
		found = false

		found = (saveConfig("interface.%d.name", i) != "") || found
		found = (saveConfig("interface.%d.mac", i) != "") || found
		found = (saveConfig("interface.%d.dhcp", i) != "") || found

		role, _ := v.readConfig(fmt.Sprintf("interface.%d.role", i))
		for a := 0; ; a++ {
			address := saveConfig("interface.%d.ip.%d.address", i, a)
			if address == "" {
				break
			} else {
				found = true
			}

			ip, _, err := net.ParseCIDR(address)
			if err != nil {
				return metadata, err
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
				return metadata, fmt.Errorf("unrecognized role: %q", role)
			}
		}

		for r := 0; ; r++ {
			gateway := saveConfig("interface.%d.route.%d.gateway", i, r)
			destination := saveConfig("interface.%d.route.%d.destination", i, r)

			if gateway == "" && destination == "" {
				break
			} else {
				found = true
			}
		}
	}
	metadata.NetworkConfig = netconf

	return
}

func (v vmware) FetchUserdata() ([]byte, error) {
	encoding, err := v.readConfig("coreos.config.data.encoding")
	if err != nil {
		return nil, err
	}

	data, err := v.readConfig("coreos.config.data")
	if err != nil {
		return nil, err
	}

	// Try to fallback to url if no explicit data
	if data == "" {
		url, err := v.readConfig("coreos.config.url")
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

func (v vmware) Type() string {
	return "vmware"
}
