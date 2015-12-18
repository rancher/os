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
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/metadata"
	"github.com/coreos/coreos-cloudinit/pkg"
)

const (
	DefaultAddress = "http://169.254.169.254/"
	apiVersion     = "2009-04-04/"
	userdataPath   = apiVersion + "user-data"
	metadataPath   = apiVersion + "meta-data"
)

type metadataService struct {
	metadata.MetadataService
}

func NewDatasource(root string) *metadataService {
	return &metadataService{metadata.NewDatasource(root, apiVersion, userdataPath, metadataPath)}
}

func (ms metadataService) FetchMetadata() (datasource.Metadata, error) {
	metadata := datasource.Metadata{}

	if keynames, err := ms.fetchAttributes(fmt.Sprintf("%s/public-keys", ms.MetadataUrl())); err == nil {
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
			sshkey, err := ms.fetchAttribute(fmt.Sprintf("%s/public-keys/%s/openssh-key", ms.MetadataUrl(), id))
			if err != nil {
				return metadata, err
			}
			metadata.SSHPublicKeys[name] = sshkey
			log.Printf("Found SSH key for %q\n", name)
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if hostname, err := ms.fetchAttribute(fmt.Sprintf("%s/hostname", ms.MetadataUrl())); err == nil {
		metadata.Hostname = strings.Split(hostname, " ")[0]
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if localAddr, err := ms.fetchAttribute(fmt.Sprintf("%s/local-ipv4", ms.MetadataUrl())); err == nil {
		metadata.PrivateIPv4 = net.ParseIP(localAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if publicAddr, err := ms.fetchAttribute(fmt.Sprintf("%s/public-ipv4", ms.MetadataUrl())); err == nil {
		metadata.PublicIPv4 = net.ParseIP(publicAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	return metadata, nil
}

func (ms metadataService) Type() string {
	return "ec2-metadata-service"
}

func (ms metadataService) fetchAttributes(url string) ([]string, error) {
	resp, err := ms.FetchData(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

func (ms metadataService) fetchAttribute(url string) (string, error) {
	if attrs, err := ms.fetchAttributes(url); err == nil && len(attrs) > 0 {
		return attrs[0], nil
	} else {
		return "", err
	}
}
