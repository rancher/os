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

package system

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
)

const DefaultIpv4Address = "127.0.0.1"

type EtcHosts struct {
	config.EtcHosts
}

func (eh EtcHosts) generateEtcHosts() (out string, err error) {
	if eh.EtcHosts != "localhost" {
		return "", errors.New("Invalid option to manage_etc_hosts")
	}

	// use the operating system hostname
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s\n", DefaultIpv4Address, hostname), nil

}

func (eh EtcHosts) File() (*File, error) {
	if eh.EtcHosts == "" {
		return nil, nil
	}

	etcHosts, err := eh.generateEtcHosts()
	if err != nil {
		return nil, err
	}

	return &File{config.File{
		Path:               path.Join("etc", "hosts"),
		RawFilePermissions: "0644",
		Content:            etcHosts,
	}}, nil
}
