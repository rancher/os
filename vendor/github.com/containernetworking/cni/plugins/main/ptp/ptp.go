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
	"net"
	"os"
	"runtime"

	"github.com/vishvananda/netlink"

	"github.com/containernetworking/cni/pkg/ip"
	"github.com/containernetworking/cni/pkg/ipam"
	"github.com/containernetworking/cni/pkg/ns"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/utils"
)

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()
}

type NetConf struct {
	types.NetConf
	IPMasq bool `json:"ipMasq"`
	MTU    int  `json:"mtu"`
}

func setupContainerVeth(netns, ifName string, mtu int, pr *types.Result) (string, error) {
	// The IPAM result will be something like IP=192.168.3.5/24, GW=192.168.3.1.
	// What we want is really a point-to-point link but veth does not support IFF_POINTOPONT.
	// Next best thing would be to let it ARP but set interface to 192.168.3.5/32 and
	// add a route like "192.168.3.0/24 via 192.168.3.1 dev $ifName".
	// Unfortunately that won't work as the GW will be outside the interface's subnet.

	// Our solution is to configure the interface with 192.168.3.5/24, then delete the
	// "192.168.3.0/24 dev $ifName" route that was automatically added. Then we add
	// "192.168.3.1/32 dev $ifName" and "192.168.3.0/24 via 192.168.3.1 dev $ifName".
	// In other words we force all traffic to ARP via the gateway except for GW itself.

	var hostVethName string
	err := ns.WithNetNSPath(netns, func(hostNS ns.NetNS) error {
		hostVeth, _, err := ip.SetupVeth(ifName, mtu, hostNS)
		if err != nil {
			return err
		}

		if err = ipam.ConfigureIface(ifName, pr); err != nil {
			return err
		}

		contVeth, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to look up %q: %v", ifName, err)
		}

		// Delete the route that was automatically added
		route := netlink.Route{
			LinkIndex: contVeth.Attrs().Index,
			Dst: &net.IPNet{
				IP:   pr.IP4.IP.IP.Mask(pr.IP4.IP.Mask),
				Mask: pr.IP4.IP.Mask,
			},
			Scope: netlink.SCOPE_NOWHERE,
		}

		if err := netlink.RouteDel(&route); err != nil {
			return fmt.Errorf("failed to delete route %v: %v", route, err)
		}

		for _, r := range []netlink.Route{
			netlink.Route{
				LinkIndex: contVeth.Attrs().Index,
				Dst: &net.IPNet{
					IP:   pr.IP4.Gateway,
					Mask: net.CIDRMask(32, 32),
				},
				Scope: netlink.SCOPE_LINK,
				Src:   pr.IP4.IP.IP,
			},
			netlink.Route{
				LinkIndex: contVeth.Attrs().Index,
				Dst: &net.IPNet{
					IP:   pr.IP4.IP.IP.Mask(pr.IP4.IP.Mask),
					Mask: pr.IP4.IP.Mask,
				},
				Scope: netlink.SCOPE_UNIVERSE,
				Gw:    pr.IP4.Gateway,
				Src:   pr.IP4.IP.IP,
			},
		} {
			if err := netlink.RouteAdd(&r); err != nil {
				return fmt.Errorf("failed to add route %v: %v", r, err)
			}
		}

		hostVethName = hostVeth.Attrs().Name

		return nil
	})
	return hostVethName, err
}

func setupHostVeth(vethName string, ipConf *types.IPConfig) error {
	// hostVeth moved namespaces and may have a new ifindex
	veth, err := netlink.LinkByName(vethName)
	if err != nil {
		return fmt.Errorf("failed to lookup %q: %v", vethName, err)
	}

	// TODO(eyakubovich): IPv6
	ipn := &net.IPNet{
		IP:   ipConf.Gateway,
		Mask: net.CIDRMask(32, 32),
	}
	addr := &netlink.Addr{IPNet: ipn, Label: ""}
	if err = netlink.AddrAdd(veth, addr); err != nil {
		return fmt.Errorf("failed to add IP addr (%#v) to veth: %v", ipn, err)
	}

	ipn = &net.IPNet{
		IP:   ipConf.IP.IP,
		Mask: net.CIDRMask(32, 32),
	}
	// dst happens to be the same as IP/net of host veth
	if err = ip.AddHostRoute(ipn, nil, veth); err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to add route on host: %v", err)
	}

	return nil
}

func cmdAdd(args *skel.CmdArgs) error {
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	if err := ip.EnableIP4Forward(); err != nil {
		return fmt.Errorf("failed to enable forwarding: %v", err)
	}

	// run the IPAM plugin and get back the config to apply
	result, err := ipam.ExecAdd(conf.IPAM.Type, args.StdinData)
	if err != nil {
		return err
	}
	if result.IP4 == nil {
		return errors.New("IPAM plugin returned missing IPv4 config")
	}

	hostVethName, err := setupContainerVeth(args.Netns, args.IfName, conf.MTU, result)
	if err != nil {
		return err
	}

	if err = setupHostVeth(hostVethName, result.IP4); err != nil {
		return err
	}

	if conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		if err = ip.SetupIPMasq(&result.IP4.IP, chain, comment); err != nil {
			return err
		}
	}

	result.DNS = conf.DNS
	return result.Print()
}

func cmdDel(args *skel.CmdArgs) error {
	conf := NetConf{}
	if err := json.Unmarshal(args.StdinData, &conf); err != nil {
		return fmt.Errorf("failed to load netconf: %v", err)
	}

	if err := ipam.ExecDel(conf.IPAM.Type, args.StdinData); err != nil {
		return err
	}

	if args.Netns == "" {
		return nil
	}

	var ipn *net.IPNet
	err := ns.WithNetNSPath(args.Netns, func(_ ns.NetNS) error {
		var err error
		ipn, err = ip.DelLinkByNameAddr(args.IfName, netlink.FAMILY_V4)
		return err
	})
	if err != nil {
		return err
	}

	if conf.IPMasq {
		chain := utils.FormatChainName(conf.Name, args.ContainerID)
		comment := utils.FormatComment(conf.Name, args.ContainerID)
		if err = ip.TeardownIPMasq(ipn, chain, comment); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	skel.PluginMain(cmdAdd, cmdDel)
}
