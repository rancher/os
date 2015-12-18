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
	"encoding/json"
	"net"
	"strconv"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/metadata"
)

const (
	DefaultAddress = "https://metadata.packet.net/"
	apiVersion     = ""
	userdataUrl    = "userdata"
	metadataPath   = "metadata"
)

type Netblock struct {
	Address       net.IP `json:"address"`
	Cidr          int    `json:"cidr"`
	Netmask       net.IP `json:"netmask"`
	Gateway       net.IP `json:"gateway"`
	AddressFamily int    `json:"address_family"`
	Public        bool   `json:"public"`
}

type Nic struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
}

type NetworkData struct {
	Interfaces []Nic      `json:"interfaces"`
	Netblocks  []Netblock `json:"addresses"`
	DNS        []net.IP   `json:"dns"`
}

// Metadata that will be pulled from the https://metadata.packet.net/metadata only. We have the opportunity to add more later.
type Metadata struct {
	Hostname    string      `json:"hostname"`
	SSHKeys     []string    `json:"ssh_keys"`
	NetworkData NetworkData `json:"network"`
}

type metadataService struct {
	metadata.MetadataService
}

func NewDatasource(root string) *metadataService {
	return &metadataService{MetadataService: metadata.NewDatasource(root, apiVersion, userdataUrl, metadataPath)}
}

func (ms *metadataService) FetchMetadata() (metadata datasource.Metadata, err error) {
	var data []byte
	var m Metadata

	if data, err = ms.FetchData(ms.MetadataUrl()); err != nil || len(data) == 0 {
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
	metadata.Hostname = m.Hostname
	metadata.SSHPublicKeys = map[string]string{}
	for i, key := range m.SSHKeys {
		metadata.SSHPublicKeys[strconv.Itoa(i)] = key
	}

	metadata.NetworkConfig = m.NetworkData

	return
}

func (ms metadataService) Type() string {
	return "packet-metadata-service"
}
