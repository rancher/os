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

package digitalocean

import (
	"encoding/json"
	"net"
	"strconv"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/metadata"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "metadata/v1"
	userdataUrl    = apiVersion + "/user-data"
	metadataPath   = apiVersion + ".json"
)

type Address struct {
	IPAddress string `json:"ip_address"`
	Netmask   string `json:"netmask"`
	Cidr      int    `json:"cidr"`
	Gateway   string `json:"gateway"`
}

type Interface struct {
	IPv4       *Address `json:"ipv4"`
	IPv6       *Address `json:"ipv6"`
	AnchorIPv4 *Address `json:"anchor_ipv4"`
	MAC        string   `json:"mac"`
	Type       string   `json:"type"`
}

type Interfaces struct {
	Public  []Interface `json:"public"`
	Private []Interface `json:"private"`
}

type DNS struct {
	Nameservers []string `json:"nameservers"`
}

type Metadata struct {
	Hostname   string     `json:"hostname"`
	Interfaces Interfaces `json:"interfaces"`
	PublicKeys []string   `json:"public_keys"`
	DNS        DNS        `json:"dns"`
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

	if len(m.Interfaces.Public) > 0 {
		if m.Interfaces.Public[0].IPv4 != nil {
			metadata.PublicIPv4 = net.ParseIP(m.Interfaces.Public[0].IPv4.IPAddress)
		}
		if m.Interfaces.Public[0].IPv6 != nil {
			metadata.PublicIPv6 = net.ParseIP(m.Interfaces.Public[0].IPv6.IPAddress)
		}
	}
	if len(m.Interfaces.Private) > 0 {
		if m.Interfaces.Private[0].IPv4 != nil {
			metadata.PrivateIPv4 = net.ParseIP(m.Interfaces.Private[0].IPv4.IPAddress)
		}
		if m.Interfaces.Private[0].IPv6 != nil {
			metadata.PrivateIPv6 = net.ParseIP(m.Interfaces.Private[0].IPv6.IPAddress)
		}
	}
	metadata.Hostname = m.Hostname
	metadata.SSHPublicKeys = map[string]string{}
	for i, key := range m.PublicKeys {
		metadata.SSHPublicKeys[strconv.Itoa(i)] = key
	}
	metadata.NetworkConfig = m

	return
}

func (ms metadataService) Type() string {
	return "digitalocean-metadata-service"
}
