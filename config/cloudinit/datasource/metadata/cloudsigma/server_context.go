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

package cloudsigma

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/rancher/os/config/cloudinit/datasource"

	"github.com/cloudsigma/cepgo"
)

const (
	userDataFieldName = "cloudinit-user-data"
)

type serverContextService struct {
	client interface {
		All() (interface{}, error)
		Key(string) (interface{}, error)
		Meta() (map[string]string, error)
		FetchRaw(string) ([]byte, error)
	}
}

func NewServerContextService() *serverContextService {
	return &serverContextService{
		client: cepgo.NewCepgo(),
	}
}

func (_ *serverContextService) IsAvailable() bool {
	productNameFile, err := os.Open("/sys/class/dmi/id/product_name")
	if err != nil {
		return false
	}
	productName := make([]byte, 10)
	_, err = productNameFile.Read(productName)

	return err == nil && string(productName) == "CloudSigma" && hasDHCPLeases()
}

func (_ *serverContextService) AvailabilityChanges() bool {
	return true
}

func (_ *serverContextService) ConfigRoot() string {
	return ""
}

func (_ *serverContextService) Type() string {
	return "server-context"
}

func (scs *serverContextService) FetchMetadata() (metadata datasource.Metadata, err error) {
	var (
		inputMetadata struct {
			Name string            `json:"name"`
			UUID string            `json:"uuid"`
			Meta map[string]string `json:"meta"`
			Nics []struct {
				Mac      string `json:"mac"`
				IPv4Conf struct {
					InterfaceType string `json:"interface_type"`
					IP            struct {
						UUID string `json:"uuid"`
					} `json:"ip"`
				} `json:"ip_v4_conf"`
				VLAN struct {
					UUID string `json:"uuid"`
				} `json:"vlan"`
			} `json:"nics"`
		}
		rawMetadata []byte
	)

	if rawMetadata, err = scs.client.FetchRaw(""); err != nil {
		return
	}

	if err = json.Unmarshal(rawMetadata, &inputMetadata); err != nil {
		return
	}

	if inputMetadata.Name != "" {
		metadata.Hostname = inputMetadata.Name
	} else {
		metadata.Hostname = inputMetadata.UUID
	}

	metadata.SSHPublicKeys = map[string]string{}
	// CloudSigma uses an empty string, rather than no string,
	// to represent the lack of a SSH key
	if key, _ := inputMetadata.Meta["ssh_public_key"]; len(key) > 0 {
		splitted := strings.Split(key, " ")
		metadata.SSHPublicKeys[splitted[len(splitted)-1]] = key
	}

	for _, nic := range inputMetadata.Nics {
		if nic.IPv4Conf.IP.UUID != "" {
			metadata.PublicIPv4 = net.ParseIP(nic.IPv4Conf.IP.UUID)
		}
		if nic.VLAN.UUID != "" {
			if localIP, err := scs.findLocalIP(nic.Mac); err == nil {
				metadata.PrivateIPv4 = localIP
			}
		}
	}

	return
}

func (scs *serverContextService) FetchUserdata() ([]byte, error) {
	metadata, err := scs.client.Meta()
	if err != nil {
		return []byte{}, err
	}

	userData, ok := metadata[userDataFieldName]
	if ok && isBase64Encoded(userDataFieldName, metadata) {
		if decodedUserData, err := base64.StdEncoding.DecodeString(userData); err == nil {
			return decodedUserData, nil
		} else {
			return []byte{}, nil
		}
	}

	return []byte(userData), nil
}

func (scs *serverContextService) findLocalIP(mac string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	ifaceMac, err := net.ParseMAC(mac)
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if !bytes.Equal(iface.HardwareAddr, ifaceMac) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch ip := addr.(type) {
			case *net.IPNet:
				if ip.IP.To4() != nil {
					return ip.IP.To4(), nil
				}
			}
		}
	}
	return nil, errors.New("Local IP not found")
}

func isBase64Encoded(field string, userdata map[string]string) bool {
	base64Fields, ok := userdata["base64_fields"]
	if !ok {
		return false
	}

	for _, base64Field := range strings.Split(base64Fields, ",") {
		if field == base64Field {
			return true
		}
	}
	return false
}

func hasDHCPLeases() bool {
	files, err := ioutil.ReadDir("/run/systemd/netif/leases/")
	return err == nil && len(files) > 0
}
