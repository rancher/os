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

package cloudstack

import (
	"net"
	"strconv"
	"strings"

	"github.com/burmilla/os/config/cloudinit/datasource"
	"github.com/burmilla/os/config/cloudinit/datasource/metadata"
	"github.com/burmilla/os/config/cloudinit/pkg"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/netconf"
)

const (
	apiVersion   = "latest/"
	userdataPath = apiVersion + "user-data"
	metadataPath = apiVersion + "meta-data/"

	serverIdentifier = "dhcp_server_identifier"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) []*MetadataService {
	roots := make([]string, 0, 5)

	if root == "" {
		if links, err := netconf.GetValidLinkList(); err == nil {
			log.Infof("Checking to see if a cloudstack server-identifier is available")
			for _, link := range links {
				linkName := link.Attrs().Name
				log.Infof("searching for cloudstack server %s on %s", serverIdentifier, linkName)
				lease := netconf.GetDhcpLease(linkName)
				if server, ok := lease[serverIdentifier]; ok {
					log.Infof("found cloudstack server '%s'", server)
					server = "http://" + server + "/"
					roots = append(roots, server)
				}
			}
		} else {
			log.Errorf("error getting LinkList: %s", err)
		}
	} else {
		roots = append(roots, root)
	}

	sources := make([]*MetadataService, 0, len(roots))
	for _, server := range roots {
		datasource := metadata.NewDatasourceWithCheckPath(server, apiVersion, metadataPath, userdataPath, metadataPath, nil)
		sources = append(sources, &MetadataService{datasource})
	}
	return sources
}

func (ms MetadataService) AvailabilityChanges() bool {
	// TODO: if it can't find the network, maybe we can start it?
	return false
}

func (ms MetadataService) FetchMetadata() (datasource.Metadata, error) {
	metadata := datasource.Metadata{}

	if sshKeys, err := ms.FetchAttributes("public-keys"); err == nil {
		metadata.SSHPublicKeys = map[string]string{}
		for i, sshkey := range sshKeys {
			log.Printf("Found SSH key %d", i)
			metadata.SSHPublicKeys[strconv.Itoa(i)] = sshkey
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if hostname, err := ms.FetchAttribute("local-hostname"); err == nil {
		metadata.Hostname = strings.Split(hostname, " ")[0]
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

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

	return metadata, nil
}

func (ms MetadataService) Type() string {
	return "cloudstack-metadata-service"
}
