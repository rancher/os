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
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/rancher/os/log"

	"github.com/docker/docker/pkg/mount"
	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/util"
)

const (
	configDevName       = "config-2"
	configDev           = "LABEL=" + configDevName
	configDevMountPoint = "/media/config-2"
	openstackAPIVersion = "latest"
)

type ConfigDrive struct {
	root                string
	readFile            func(filename string) ([]byte, error)
	lastError           error
	availabilityChanges bool
}

func NewDatasource(root string) *ConfigDrive {
	return &ConfigDrive{root, ioutil.ReadFile, nil, true}
}

func (cd *ConfigDrive) IsAvailable() bool {
	if cd.root == configDevMountPoint {
		cd.lastError = MountConfigDrive()
		if cd.lastError != nil {
			log.Error(cd.lastError)
			// Don't keep retrying if we can't mount
			cd.availabilityChanges = false
			return false
		}
		defer cd.Finish()
	}

	_, cd.lastError = os.Stat(cd.root)
	return !os.IsNotExist(cd.lastError)
	// TODO: consider changing IsNotExists to not-available _and_ does not change
}

func (cd *ConfigDrive) Finish() error {
	return UnmountConfigDrive()
}

func (cd *ConfigDrive) String() string {
	if cd.lastError != nil {
		return fmt.Sprintf("%s: %s (lastError: %s)", cd.Type(), cd.root, cd.lastError)
	}
	return fmt.Sprintf("%s: %s", cd.Type(), cd.root)
}

func (cd *ConfigDrive) AvailabilityChanges() bool {
	return cd.availabilityChanges
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
	// TODO: I don't think we've used this for anything
	/*	if m.NetworkConfig.ContentPath != "" {
			metadata.NetworkConfig, err = cd.tryReadFile(path.Join(cd.openstackRoot(), m.NetworkConfig.ContentPath))
		}
	*/
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
	if cd.root == configDevMountPoint {
		cd.lastError = MountConfigDrive()
		if cd.lastError != nil {
			log.Error(cd.lastError)
			return nil, cd.lastError
		}
		defer cd.Finish()
	}
	log.Debugf("Attempting to read from %q\n", filename)
	data, err := cd.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		log.Errorf("ERROR read cloud-config file(%s) - err: %q", filename, err)
	}
	return data, err
}

func MountConfigDrive() error {
	if err := os.MkdirAll(configDevMountPoint, 700); err != nil {
		return err
	}

	configDev := util.ResolveDevice(configDev)

	if configDev == "" {
		return mount.Mount(configDevName, configDevMountPoint, "9p", "trans=virtio,version=9p2000.L")
	}

	fsType, err := util.GetFsType(configDev)
	if err != nil {
		return err
	}
	return mount.Mount(configDev, configDevMountPoint, fsType, "ro")
}

func UnmountConfigDrive() error {
	return syscall.Unmount(configDevMountPoint, 0)
}
