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

package configdrive

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/rancher/os/config/cloudinit/datasource"
)

const (
	openstackAPIVersion = "latest"
)

type ConfigDrive struct {
	root     string
	readFile func(filename string) ([]byte, error)
}

func NewDatasource(root string) *ConfigDrive {
	return &ConfigDrive{root, ioutil.ReadFile}
}

func (cd *ConfigDrive) IsAvailable() bool {
	_, err := os.Stat(cd.root)
	return !os.IsNotExist(err)
}

func (cd *ConfigDrive) AvailabilityChanges() bool {
	return true
}

func (cd *ConfigDrive) ConfigRoot() string {
	return cd.openstackRoot()
}

func (cd *ConfigDrive) FetchMetadata() (metadata datasource.Metadata, err error) {
	var data []byte
	var m struct {
		SSHAuthorizedKeyMap map[string]string `json:"public_keys"`
		Hostname            string            `json:"hostname"`
		NetworkConfig       struct {
			ContentPath string `json:"content_path"`
		} `json:"network_config"`
	}

	if data, err = cd.tryReadFile(path.Join(cd.openstackVersionRoot(), "meta_data.json")); err != nil || len(data) == 0 {
		return
	}
	if err = json.Unmarshal([]byte(data), &m); err != nil {
		return
	}

	metadata.SSHPublicKeys = m.SSHAuthorizedKeyMap
	metadata.Hostname = m.Hostname
	if m.NetworkConfig.ContentPath != "" {
		metadata.NetworkConfig, err = cd.tryReadFile(path.Join(cd.openstackRoot(), m.NetworkConfig.ContentPath))
	}

	return
}

func (cd *ConfigDrive) FetchUserdata() ([]byte, error) {
	return cd.tryReadFile(path.Join(cd.openstackVersionRoot(), "user_data"))
}

func (cd *ConfigDrive) Type() string {
	return "cloud-drive"
}

func (cd *ConfigDrive) openstackRoot() string {
	return path.Join(cd.root, "openstack")
}

func (cd *ConfigDrive) openstackVersionRoot() string {
	return path.Join(cd.openstackRoot(), openstackAPIVersion)
}

func (cd *ConfigDrive) tryReadFile(filename string) ([]byte, error) {
	log.Printf("Attempting to read from %q\n", filename)
	data, err := cd.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}
