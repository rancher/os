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

package ec2

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/rancher/os/netconf"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/config/cloudinit/pkg"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "latest/"
	userdataPath   = apiVersion + "user-data"
	metadataPath   = apiVersion + "meta-data/"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) *MetadataService {
	if root == "" {
		root = DefaultAddress
	}
	return &MetadataService{metadata.NewDatasource(root, apiVersion, userdataPath, metadataPath, nil)}
}

func (ms MetadataService) AvailabilityChanges() bool {
	// TODO: if it can't find the network, maybe we can start it?
	return false
}

func (ms MetadataService) FetchMetadata() (datasource.Metadata, error) {
	// see http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html
	metadata := datasource.Metadata{}
	metadata.NetworkConfig = netconf.NetworkConfig{}

	if keynames, err := ms.FetchAttributes("public-keys"); err == nil {
		keyIDs := make(map[string]string)
		for _, keyname := range keynames {
			tokens := strings.SplitN(keyname, "=", 2)
			if len(tokens) != 2 {
				return metadata, fmt.Errorf("malformed public key: %q", keyname)
			}
			keyIDs[tokens[1]] = tokens[0]
		}

		metadata.SSHPublicKeys = map[string]string{}
		for name, id := range keyIDs {
			sshkey, err := ms.FetchAttribute(fmt.Sprintf("public-keys/%s/openssh-key", id))
			if err != nil {
				return metadata, err
			}
			metadata.SSHPublicKeys[name] = sshkey
			log.Printf("Found SSH key for %q\n", name)
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if hostname, err := ms.FetchAttribute("hostname"); err == nil {
		metadata.Hostname = strings.Split(hostname, " ")[0]
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	// TODO: these are only on the first interface - it looks like you can have as many as you need...
	if localAddr, err := ms.FetchAttribute("local-ipv4"); err == nil {
		metadata.PrivateIPv4 = net.ParseIP(localAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}
	if publicAddr, err := ms.FetchAttribute("public-ipv4"); err == nil {
		metadata.PublicIPv4 = net.ParseIP(publicAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	metadata.NetworkConfig.Interfaces = make(map[string]netconf.InterfaceConfig)
	if macs, err := ms.FetchAttributes("network/interfaces/macs"); err != nil {
		for _, mac := range macs {
			if deviceNumber, err := ms.FetchAttribute(fmt.Sprintf("network/interfaces/macs/%s/device-number", mac)); err != nil {
				network := netconf.InterfaceConfig{
					DHCP: true,
				}
				/* Looks like we must use DHCP for aws
				// private ipv4
				if subnetCidrBlock, err := ms.FetchAttribute(fmt.Sprintf("network/interfaces/macs/%s/subnet-ipv4-cidr-block", mac)); err != nil {
					cidr := strings.Split(subnetCidrBlock, "/")
					if localAddr, err := ms.FetchAttributes(fmt.Sprintf("network/interfaces/macs/%s/local-ipv4s", mac)); err != nil {
						for _, addr := range localAddr {
							network.Addresses = append(network.Addresses, addr+"/"+cidr[1])
						}
					}
				}
				// ipv6
				if localAddr, err := ms.FetchAttributes(fmt.Sprintf("network/interfaces/macs/%s/ipv6s", mac)); err != nil {
					if subnetCidrBlock, err := ms.FetchAttributes(fmt.Sprintf("network/interfaces/macs/%s/subnet-ipv6-cidr-block", mac)); err != nil {
						for i, addr := range localAddr {
							cidr := strings.Split(subnetCidrBlock[i], "/")
							network.Addresses = append(network.Addresses, addr+"/"+cidr[1])
						}
					}
				}
				*/
				// disabled - it looks to me like you don't actually put the public IP on the eth device
				/*				if publicAddr, err := ms.FetchAttributes(fmt.Sprintf("network/interfaces/macs/%s/public-ipv4s", mac)); err != nil {
									if vpcCidrBlock, err := ms.FetchAttribute(fmt.Sprintf("network/interfaces/macs/%s/vpc-ipv4-cidr-block", mac)); err != nil {
										cidr := strings.Split(vpcCidrBlock, "/")
										network.Addresses = append(network.Addresses, publicAddr+"/"+cidr[1])
									}
								}
				*/

				metadata.NetworkConfig.Interfaces["eth"+deviceNumber] = network
			}
		}
	}

	return metadata, nil
}

func (ms MetadataService) Type() string {
	return "ec2-metadata-service"
}
