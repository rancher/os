// Copyright 2015 CNI authors
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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/containernetworking/cni/pkg/ip"
	"github.com/containernetworking/cni/pkg/ipam"
	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/vishvananda/netlink"
)

type NetConf struct {
	types.NetConf
	Master string `json:"master"`
	Mode   string `json:"mode"`
	MTU    int    `json:"mtu"`
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

func loadConf(bytes []byte) (*NetConf, error) {
	n := &NetConf{}
	if err := json.Unmarshal(bytes, n); err != nil {
		return nil, fmt.Errorf("failed to load netconf: %v", err)
	}
	if n.Master == "" {
		return nil, fmt.Errorf(`"master" field is required. It specifies the host interface name to virtualize`)
	}
	return n, nil
}

func modeFromString(s string) (netlink.IPVlanMode, error) {
	switch s {
	case "", "l2":
		return netlink.IPVLAN_MODE_L2, nil
	case "l3":
		return netlink.IPVLAN_MODE_L3, nil
	default:
		return 0, fmt.Errorf("unknown ipvlan mode: %q", s)
	}
}

func createIpvlan(conf *NetConf, ifName string, netns ns.NetNS) error {
	mode, err := modeFromString(conf.Mode)
	if err != nil {
		return err
	}

	m, err := netlink.LinkByName(conf.Master)
	if err != nil {
		return fmt.Errorf("failed to lookup master %q: %v", conf.Master, err)
	}

	// due to kernel bug we have to create with tmpname or it might
	// collide with the name on the host and error out
	tmpName, err := ip.RandomVethName()
	if err != nil {
		return err
	}

	mv := &netlink.IPVlan{
		LinkAttrs: netlink.LinkAttrs{
			MTU:         conf.MTU,
			Name:        tmpName,
			ParentIndex: m.Attrs().Index,
			Namespace:   netlink.NsFd(int(netns.Fd())),
		},
		Mode: mode,
	}

	if err := netlink.LinkAdd(mv); err != nil {
		return fmt.Errorf("failed to create ipvlan: %v", err)
	}

	return netns.Do(func(_ ns.NetNS) error {
		err := renameLink(tmpName, ifName)
		if err != nil {
			return fmt.Errorf("failed to rename ipvlan to %q: %v", ifName, err)
		}
		return nil
	})
}

func cmdAdd(args *skel.CmdArgs) error {
	n, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	netns, err := ns.GetNS(args.Netns)
	if err != nil {
		return fmt.Errorf("failed to open netns %q: %v", args.Netns, err)
	}
	defer netns.Close()

	if err = createIpvlan(n, args.IfName, netns); err != nil {
		return err
	}

	// run the IPAM plugin and get back the config to apply
	result, err := ipam.ExecAdd(n.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	if result.IP4 == nil {
		return errors.New("IPAM plugin returned missing IPv4 config")
	}

	err = netns.Do(func(_ ns.NetNS) error {
		return ipam.ConfigureIface(args.IfName, result)
	})
	if err != nil {
		return err
	}

	result.DNS = n.DNS
	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	n, err := loadConf(args.StdinData)
	if err != nil {
		return err
	}

	err = ipam.ExecDel(n.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	return ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		return ip.DelLinkByName(args.IfName)
	})
}

func renameLink(curName, newName string) error {
	link, err := netlink.LinkByName(curName)
	if err != nil {
		return err
	}

	return netlink.LinkSetName(link, newName)
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel)
}
