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

package packet

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/log"
	"github.com/rancher/os/netconf"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	packetMetadata "github.com/packethost/packngo/metadata"
)

const (
	DefaultAddress = "https://metadata.packet.net/"
	apiVersion     = ""
	userdataURL    = "userdata"
	metadataPath   = "metadata"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) *MetadataService {
	if root == "" {
		root = DefaultAddress
	}

	return &MetadataService{Service: metadata.NewDatasource(root, apiVersion, userdataURL, metadataPath, nil)}
}

func (ms *MetadataService) FetchMetadata() (metadata datasource.Metadata, err error) {
	m, err := packetMetadata.GetMetadata()
	if err != nil {
		log.Errorf("Failed to get Packet metadata: %v", err)
		return
	}

	bondCfg := netconf.InterfaceConfig{
		Addresses: []string{},
		BondOpts: map[string]string{
			"lacp_rate":        "1",
			"xmit_hash_policy": "layer3+4",
			"downdelay":        "200",
			"updelay":          "200",
			"miimon":           "100",
			"mode":             "4",
		},
	}
	netCfg := netconf.NetworkConfig{
		Interfaces: map[string]netconf.InterfaceConfig{},
	}
	for _, iface := range m.Network.Interfaces {
		netCfg.Interfaces["mac="+iface.MAC] = netconf.InterfaceConfig{
			Bond: "bond0",
		}
	}
	for _, addr := range m.Network.Addresses {
		bondCfg.Addresses = append(bondCfg.Addresses, fmt.Sprintf("%s/%d", addr.Address, addr.NetworkBits))
		if addr.Gateway != nil && len(addr.Gateway) > 0 {
			if addr.Family == packetMetadata.IPv4 {
				if addr.Public {
					bondCfg.Gateway = addr.Gateway.String()
				}
			} else {
				bondCfg.GatewayIpv6 = addr.Gateway.String()
			}
		}

		if addr.Family == packetMetadata.IPv4 && strings.HasPrefix(addr.Gateway.String(), "10.") {
			bondCfg.PostUp = append(bondCfg.PostUp, "ip route add 10.0.0.0/8 via "+addr.Gateway.String())
		}
	}

	netCfg.Interfaces["bond0"] = bondCfg
	b, _ := yaml.Marshal(netCfg)
	log.Debugf("Generated network config: %s", string(b))

	// the old code	var data []byte
	/*	var m Metadata

		if data, err = ms.FetchData(ms.MetadataURL()); err != nil || len(data) == 0 {
			return
		}

		if err = json.Unmarshal(data, &m); err != nil {
			return
		}

		if len(m.NetworkData.Netblocks) > 0 {
			for _, Netblock := range m.NetworkData.Netblocks {
				if Netblock.AddressFamily == 4 {
					if Netblock.Public == true {
						metadata.PublicIPv4 = Netblock.Address
					} else {
						metadata.PrivateIPv4 = Netblock.Address
					}
				} else {
					metadata.PublicIPv6 = Netblock.Address
				}
			}
		}
	*/
	metadata.Hostname = m.Hostname
	metadata.SSHPublicKeys = map[string]string{}
	for i, key := range m.SSHKeys {
		metadata.SSHPublicKeys[strconv.Itoa(i)] = key
	}

	metadata.NetworkConfig = netCfg

	// This is not really the right place - perhaps we should add a call-home function in each datasource to be called after the network is applied
	//(see the original in cmd/cloudsave/packet)
	//if _, err = http.Post(m.PhoneHomeURL, "application/json", bytes.NewReader([]byte{})); err != nil {
	//log.Errorf("Failed to post to Packet phone home URL: %v", err)
	//}

	return
}

func (ms MetadataService) Type() string {
	return "packet-metadata-service"
}
