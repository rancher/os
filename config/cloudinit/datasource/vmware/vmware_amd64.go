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
	"io/ioutil"
	"os"

	"github.com/burmilla/os/config/cloudinit/pkg"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"

	"github.com/sigma/vmw-guestinfo/rpcvmx"
	"github.com/sigma/vmw-guestinfo/vmcheck"
	ovf "github.com/sigma/vmw-ovflib"
)

type ovfWrapper struct {
	env *ovf.OvfEnvironment
}

func (ovf ovfWrapper) readConfig(key string) (string, error) {
	if val := ovf.env.Properties["guestinfo."+key]; val != "" {
		return val, nil
	}
	return readConfig(key)
}

func NewDatasource(fileName string) *VMWare {
	if util.GetHypervisor() != "vmware" {
		return nil
	}
	// read from provided ovf environment document (typically /media/ovfenv/ovf-env.xml)
	if fileName != "" {
		log.Printf("Using OVF environment from %s\n", fileName)
		ovfEnv, err := ioutil.ReadFile(fileName)
		if err != nil {
			ovfEnv = make([]byte, 0)
		}
		return &VMWare{
			ovfFileName: fileName,
			readConfig:  getOvfReadConfig(ovfEnv),
			urlDownload: urlDownload,
		}
	}

	// try to read ovf environment from VMware tools
	data, err := readConfig("ovfenv")
	if err == nil && data != "" {
		log.Printf("Using OVF environment from guestinfo\n")
		return &VMWare{
			readConfig:  getOvfReadConfig([]byte(data)),
			urlDownload: urlDownload,
		}
	}

	// if everything fails, fallback to directly reading variables from the backdoor
	log.Printf("Using guestinfo variables\n")
	return &VMWare{
		readConfig:  readConfig,
		urlDownload: urlDownload,
	}
}

func (v VMWare) IsAvailable() bool {
	if util.GetHypervisor() != "vmware" {
		return false
	}
	if v.ovfFileName != "" {
		_, v.lastError = os.Stat(v.ovfFileName)
		if !os.IsNotExist(v.lastError) {
			// when GuestInfo is empty, the DataSource should not be available.
			return v.checkGuestInfo()
		}
		return false
	}
	// when GuestInfo is empty, the DataSource should not be available.
	return vmcheck.IsVirtualWorld() && v.checkGuestInfo()
}

func readConfig(key string) (string, error) {
	data, err := rpcvmx.NewConfig().String(key, "")
	if err == nil {
		log.Printf("Read from %q: %q\n", key, data)
	} else {
		log.Printf("Failed to read from %q: %v\n", key, err)
	}
	return data, err
}

func getOvfReadConfig(ovfEnv []byte) readConfigFunction {
	env := &ovf.OvfEnvironment{}
	if len(ovfEnv) != 0 {
		env = ovf.ReadEnvironment(ovfEnv)
	}

	wrapper := ovfWrapper{env}
	return wrapper.readConfig
}

func urlDownload(url string) ([]byte, error) {
	client := pkg.NewHTTPClient()
	return client.GetRetry(url)
}

func (v VMWare) checkGuestInfo() bool {
	userData, err := v.FetchUserdata()
	if err == nil && string(userData) != "" {
		return true
	}
	metadata, err := v.FetchMetadata()
	if err == nil {
		if metadata.Hostname != "" {
			return true
		}
		if len(metadata.NetworkConfig.DNS.Nameservers) > 0 || len(metadata.NetworkConfig.DNS.Search) > 0 {
			return true
		}
		if len(metadata.NetworkConfig.Interfaces) > 0 {
			return true
		}
	}
	return false
}
