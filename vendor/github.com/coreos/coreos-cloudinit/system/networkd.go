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
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/network"

	"github.com/docker/libcontainer/netlink"
)

func RestartNetwork(interfaces []network.InterfaceGenerator) (err error) {
	defer func() {
		if e := restartNetworkd(); e != nil {
			err = e
		}
	}()

	if err = downNetworkInterfaces(interfaces); err != nil {
		return
	}

	if err = maybeProbe8012q(interfaces); err != nil {
		return
	}
	return maybeProbeBonding(interfaces)
}

func downNetworkInterfaces(interfaces []network.InterfaceGenerator) error {
	sysInterfaceMap := make(map[string]*net.Interface)
	if systemInterfaces, err := net.Interfaces(); err == nil {
		for _, iface := range systemInterfaces {
			iface := iface
			sysInterfaceMap[iface.Name] = &iface
		}
	} else {
		return err
	}

	for _, iface := range interfaces {
		if systemInterface, ok := sysInterfaceMap[iface.Name()]; ok {
			log.Printf("Taking down interface %q\n", systemInterface.Name)
			if err := netlink.NetworkLinkDown(systemInterface); err != nil {
				fmt.Printf("Error while downing interface %q (%s). Continuing...\n", systemInterface.Name, err)
			}
		}
	}

	return nil
}

func maybeProbe8012q(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		if iface.Type() == "vlan" {
			log.Printf("Probing LKM %q (%q)\n", "8021q", "8021q")
			return exec.Command("modprobe", "8021q").Run()
		}
	}
	return nil
}

func maybeProbeBonding(interfaces []network.InterfaceGenerator) error {
	for _, iface := range interfaces {
		if iface.Type() == "bond" {
			args := append([]string{"bonding"}, strings.Split(iface.ModprobeParams(), " ")...)
			log.Printf("Probing LKM %q (%q)\n", "bonding", args)
			return exec.Command("modprobe", args...).Run()
		}
	}
	return nil
}

func restartNetworkd() error {
	log.Printf("Restarting networkd.service\n")
	networkd := Unit{config.Unit{Name: "systemd-networkd.service"}}
	_, err := NewUnitManager("").RunUnitCommand(networkd, "restart")
	return err
}
