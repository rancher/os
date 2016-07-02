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

package waagent

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/datasource"
)

type waagent struct {
	root     string
	readFile func(filename string) ([]byte, error)
}

func NewDatasource(root string) *waagent {
	return &waagent{root, ioutil.ReadFile}
}

func (a *waagent) IsAvailable() bool {
	_, err := os.Stat(path.Join(a.root, "provisioned"))
	return !os.IsNotExist(err)
}

func (a *waagent) AvailabilityChanges() bool {
	return true
}

func (a *waagent) ConfigRoot() string {
	return a.root
}

func (a *waagent) unmarshalXML(filename string, v interface{}) (err error) {
	var b []byte
	if b, err = a.tryReadFile(filename); err != nil || len(b) == 0 {
		return
	}

	err = xml.Unmarshal(b, &v)
	return
}

func (a *waagent) FetchMetadata() (metadata datasource.Metadata, err error) {
	var ovf OvfEnv
	if err = a.unmarshalXML(path.Join(a.root, "ovf-env.xml"), &ovf); err != nil {
		return
	}

	metadata.Hostname = ovf.Config.Hostname

	var sc SharedConfig
	if err = a.unmarshalXML(path.Join(a.root, "SharedConfig.xml"), &sc); err != nil {
		return
	}

	var instance Instance
	for _, i := range sc.Instances.Instances {
		if i.Id == sc.Incarnation.Instance {
			instance = i
			break
		}
	}

	metadata.PrivateIPv4 = net.ParseIP(instance.Address)
	for _, e := range instance.InputEndpoints.Endpoints {
		host, _, err := net.SplitHostPort(e.LoadBalancedPublicAddress)
		if err == nil {
			metadata.PublicIPv4 = net.ParseIP(host)
			break
		}
	}
	return
}

func (a *waagent) FetchUserdata() ([]byte, error) {
	return a.tryReadFile(path.Join(a.root, "CustomData"))
}

func (a *waagent) Type() string {
	return "waagent"
}

func (a *waagent) tryReadFile(filename string) ([]byte, error) {
	log.Printf("Attempting to read from %q\n", filename)
	data, err := a.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	return data, err
}

type Instance struct {
	Id             string `xml:"id,attr"`
	Address        string `xml:"address,attr"`
	InputEndpoints struct {
		Endpoints []struct {
			LoadBalancedPublicAddress string `xml:"loadBalancedPublicAddress,attr"`
		} `xml:"Endpoint"`
	}
}

type SharedConfig struct {
	Incarnation struct {
		Instance string `xml:"instance,attr"`
	}
	Instances struct {
		Instances []Instance `xml:"Instance"`
	}
}

type OvfEnv struct {
	Config struct {
		Hostname string `xml:"HostName"`
		Username string `xml:"UserName"`
		Password string `xml:"UserPassword"`
		Userdata string `xml:"CustomData"`
	} `xml:"ProvisioningSection>LinuxProvisioningConfigurationSet"`
}
