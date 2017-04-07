package netconf

import (
	"bytes"
	"errors"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	shlex "github.com/flynn/go-shlex"
	"github.com/rancher/os/log"

	glob "github.com/ryanuber/go-glob"
	"github.com/vishvananda/netlink"
)

const (
	CONF = "/var/lib/rancher/conf"
	MODE = "mode"
)

var (
	defaultDhcpArgs = []string{"dhcpcd", "-MA4"}
	dhcpReleaseCmd  = "dhcpcd --release"
)

func createInterfaces(netCfg *NetworkConfig) {
	configured := map[string]bool{}

	for name, iface := range netCfg.Interfaces {
		if iface.Bridge == "true" {
			if _, err := NewBridge(name); err != nil {
				log.Errorf("Failed to create bridge %s: %v", name, err)
			}
		} else if iface.Bridge != "" {
			if _, err := NewBridge(iface.Bridge); err != nil {
				log.Errorf("Failed to create bridge %s: %v", iface.Bridge, err)
			}
		} else if iface.Bond != "" {
			bond, err := Bond(iface.Bond)
			if err != nil {
				log.Errorf("Failed to create bond %s: %v", iface.Bond, err)
				continue
			}

			if !configured[iface.Bond] {
				if bondIface, ok := netCfg.Interfaces[iface.Bond]; ok {
					// Other settings depends on mode, so set it first
					if v, ok := bondIface.BondOpts[MODE]; ok {
						bond.Opt(MODE, v)
					}

					for k, v := range bondIface.BondOpts {
						if k != MODE {
							bond.Opt(k, v)
						}
					}
					configured[iface.Bond] = true
				}
			}
		}
	}
}

func createSlaveInterfaces(netCfg *NetworkConfig) {
	links, err := netlink.LinkList()
	if err != nil {
		log.Errorf("Failed to list links: %v", err)
		return
	}

	for _, link := range links {
		match, ok := findMatch(link, netCfg)
		if !ok {
			continue
		}

		vlanDefs, err := ParseVlanDefinitions(match.Vlans)
		if err != nil {
			log.Errorf("Failed to create vlans on device %s: %v", link.Attrs().Name, err)
			continue
		}

		for _, vlanDef := range vlanDefs {
			if _, err = NewVlan(link, vlanDef.Name, vlanDef.ID); err != nil {
				log.Errorf("Failed to create vlans on device %s, id %d: %v", link.Attrs().Name, vlanDef.ID, err)
			}
		}
	}
}

func findMatch(link netlink.Link, netCfg *NetworkConfig) (InterfaceConfig, bool) {
	linkName := link.Attrs().Name
	var match InterfaceConfig
	exactMatch := false
	found := false

	for key, netConf := range netCfg.Interfaces {
		if netConf.Match == "" {
			netConf.Match = key
		}

		if netConf.Match == "" {
			continue
		}

		if strings.HasPrefix(netConf.Match, "mac") {
			haAddr, err := net.ParseMAC(netConf.Match[4:])
			if err != nil {
				log.Errorf("Failed to parse mac %s: %v", netConf.Match[4:], err)
				continue
			}

			// Don't match mac address of the bond because it is the same as the slave
			if bytes.Compare(haAddr, link.Attrs().HardwareAddr) == 0 && link.Attrs().Name != netConf.Bond {
				// MAC address match is used over all other matches
				return netConf, true
			}
		}

		if !exactMatch && glob.Glob(netConf.Match, linkName) {
			match = netConf
			found = true
		}

		if netConf.Match == linkName {
			// Found exact match, use it over wildcard match
			match = netConf
			exactMatch = true
		}
	}

	return match, exactMatch || found
}

func populateDefault(netCfg *NetworkConfig) {
	if netCfg.Interfaces == nil {
		netCfg.Interfaces = map[string]InterfaceConfig{}
	}

	if len(netCfg.Interfaces) == 0 {
		netCfg.Interfaces["eth*"] = InterfaceConfig{
			DHCP: true,
		}
	}

	if _, ok := netCfg.Interfaces["lo"]; !ok {
		netCfg.Interfaces["lo"] = InterfaceConfig{
			Addresses: []string{
				"127.0.0.1/8",
				"::1/128",
			},
		}
	}
}

func ApplyNetworkConfigs(netCfg *NetworkConfig) error {
	populateDefault(netCfg)

	log.Debugf("Config: %#v", netCfg)
	runCmds(netCfg.PreCmds, "")

	createInterfaces(netCfg)

	createSlaveInterfaces(netCfg)

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	//apply network config
	for _, link := range links {
		linkName := link.Attrs().Name
		if match, ok := findMatch(link, netCfg); ok && !match.DHCP {
			if err := applyInterfaceConfig(link, match); err != nil {
				log.Errorf("Failed to apply settings to %s : %v", linkName, err)
			}
		}
	}

	runCmds(netCfg.PostCmds, "")
	return err
}

func RunDhcp(netCfg *NetworkConfig, setHostname, setDNS bool) error {
	log.Debugf("RunDhcp")
	populateDefault(netCfg)

	links, err := netlink.LinkList()
	if err != nil {
		log.Errorf("RunDhcp failed to get LinkList, %s", err)
		return err
	}

	wg := sync.WaitGroup{}

	for _, link := range links {
		name := link.Attrs().Name
		if name == "lo" {
			continue
		}
		match, ok := findMatch(link, netCfg)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(iface string, match InterfaceConfig) {
			if match.DHCP {
				runDhcp(netCfg, iface, match.DHCPArgs, setHostname, setDNS)
			} else {
				runDhcp(netCfg, iface, dhcpReleaseCmd, false, true)
			}
			wg.Done()
		}(name, match)
	}
	wg.Wait()

	return nil
}

func runDhcp(netCfg *NetworkConfig, iface string, argstr string, setHostname, setDNS bool) {
	args := []string{}
	if argstr != "" {
		var err error
		args, err = shlex.Split(argstr)
		if err != nil {
			log.Errorf("Failed to parse [%s]: %v", argstr, err)
		}
	}
	if len(args) == 0 {
		args = defaultDhcpArgs
	}

	if setHostname {
		args = append(args, "-e", "force_hostname=true")
	}

	if !setDNS {
		args = append(args, "--nohook", "resolv.conf")
	}

	args = append(args, iface)
	cmd := exec.Command(args[0], args[1:]...)
	log.Infof("Running DHCP on %s: %s", iface, strings.Join(args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
}

func linkUp(link netlink.Link, netConf InterfaceConfig) error {
	if err := netlink.LinkSetUp(link); err != nil {
		log.Errorf("failed to setup link: %v", err)
		return err
	}

	return nil
}

func applyAddress(address string, link netlink.Link, netConf InterfaceConfig) error {
	addr, err := netlink.ParseAddr(address)
	if err != nil {
		return err
	}
	if err := netlink.AddrAdd(link, addr); err == syscall.EEXIST {
		//Ignore this error
	} else if err != nil {
		log.Errorf("addr add failed: %v", err)
	} else {
		log.Infof("Set %s on %s", netConf.Address, link.Attrs().Name)
	}

	return nil
}

func removeAddress(addr netlink.Addr, link netlink.Link) error {
	if err := netlink.AddrDel(link, &addr); err == syscall.EEXIST {
		//Ignore this error
	} else if err != nil {
		log.Errorf("addr del failed: %v", err)
	} else {
		log.Infof("Removed %s from %s", addr.String(), link.Attrs().Name)
	}

	return nil
}

// setGateway(add=false) will set _one_ gateway on an interface (ie, replace an existing one)
// setGateway(add=true) will add another gateway to an interface
func setGateway(gateway string, add bool) error {
	if gateway == "" {
		return nil
	}

	gatewayIP := net.ParseIP(gateway)
	if gatewayIP == nil {
		return errors.New("Invalid gateway address " + gateway)
	}

	route := netlink.Route{
		Scope: netlink.SCOPE_UNIVERSE,
		Gw:    gatewayIP,
	}

	if add {
		if err := netlink.RouteAdd(&route); err == syscall.EEXIST {
			//Ignore this error
		} else if err != nil {
			log.Errorf("gateway add failed: %v", err)
			return err
		}
		log.Infof("Added default gateway %s", gateway)
	} else {
		if err := netlink.RouteReplace(&route); err == syscall.EEXIST {
			//Ignore this error
		} else if err != nil {
			log.Errorf("gateway replace failed: %v", err)
			return err
		}
		log.Infof("Replaced default gateway %s", gateway)
	}

	return nil
}

func applyInterfaceConfig(link netlink.Link, netConf InterfaceConfig) error {
	//TODO: skip doing anything if the settings are "default"?
	//TODO: how do you undo a non-default with a default?
	// ATM, this removes

	// TODO: undo
	if netConf.Bond != "" {
		if err := netlink.LinkSetDown(link); err != nil {
			return err
		}

		b, err := Bond(netConf.Bond)
		if err != nil {
			return err
		}
		if err := b.AddSlave(link.Attrs().Name); err != nil {
			return err
		}
		return nil
	}

	//TODO: undo
	if netConf.Bridge != "" && netConf.Bridge != "true" {
		b, err := NewBridge(netConf.Bridge)
		if err != nil {
			return err
		}
		if err := b.AddLink(link); err != nil {
			return err
		}
		return linkUp(link, netConf)
	}

	if netConf.IPV4LL {
		if err := AssignLinkLocalIP(link); err != nil {
			log.Errorf("IPV4LL set failed: %v", err)
			return err
		}
	} else {
		if err := RemoveLinkLocalIP(link); err != nil {
			log.Errorf("IPV4LL del failed: %v", err)
			return err
		}
	}

	addresses := []string{}

	if netConf.Address != "" {
		addresses = append(addresses, netConf.Address)
	}

	if len(netConf.Addresses) > 0 {
		addresses = append(addresses, netConf.Addresses...)
	}

	existingAddrs, _ := getLinkAddrs(link)
	addrMap := make(map[string]bool)
	for _, address := range addresses {
		addrMap[address] = true
		log.Infof("Applying %s to %s", address, link.Attrs().Name)
		err := applyAddress(address, link, netConf)
		if err != nil {
			log.Errorf("Failed to apply address %s to %s: %v", address, link.Attrs().Name, err)
		}
	}
	for _, addr := range existingAddrs {
		if _, ok := addrMap[addr.IPNet.String()]; !ok {
			if netConf.DHCP || netConf.IPV4LL {
				// let the dhcpcd take care of it
				log.Infof("leaving  %s from %s", addr.String(), link.Attrs().Name)
			} else {
				log.Infof("removing  %s from %s", addr.String(), link.Attrs().Name)
				removeAddress(addr, link)
			}
		}
	}

	// TODO: can we set to default?
	if netConf.MTU > 0 {
		if err := netlink.LinkSetMTU(link, netConf.MTU); err != nil {
			log.Errorf("set MTU Failed: %v", err)
			return err
		}
	}

	runCmds(netConf.PreUp, link.Attrs().Name)

	if err := linkUp(link, netConf); err != nil {
		return err
	}

	// replace the existing gw with the main ipv4 one
	if err := setGateway(netConf.Gateway, true); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.Gateway)
	}
	//and then add the ipv6 one if it exists
	if err := setGateway(netConf.GatewayIpv6, true); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.GatewayIpv6)
	}

	// TODO: how to remove a GW? (on aws it seems to be hard to find out what the gw is :/)
	runCmds(netConf.PostUp, link.Attrs().Name)

	return nil
}

func runCmds(cmds []string, iface string) {
	for _, cmd := range cmds {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		args, err := shlex.Split(strings.Replace(cmd, "$iface", iface, -1))
		if err != nil {
			log.Errorf("Failed to parse command [%s]: %v", cmd, err)
			continue
		}

		log.Infof("Running command %s %v", args[0], args[1:])
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Failed to run command [%v]: %v", cmd, err)
			continue
		}
	}
}
