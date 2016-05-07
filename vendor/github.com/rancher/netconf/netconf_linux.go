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

	log "github.com/Sirupsen/logrus"
	"github.com/flynn/go-shlex"

	"github.com/ryanuber/go-glob"
	"github.com/vishvananda/netlink"
)

const (
	CONF = "/var/lib/rancher/conf"
	MODE = "mode"
)

var (
	defaultDhcpArgs = []string{"dhcpcd", "-MA4"}
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
			if _, err = NewVlan(link, vlanDef.Name, vlanDef.Id); err != nil {
				log.Errorf("Failed to create vlans on device %s, id %d: %v", link.Attrs().Name, vlanDef.Id, err)
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
			Address: "127.0.0.1/8",
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

func RunDhcp(netCfg *NetworkConfig, setHostname, setDns bool) error {
	populateDefault(netCfg)

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	dhcpLinks := map[string]string{}
	for _, link := range links {
		if match, ok := findMatch(link, netCfg); ok && match.DHCP {
			dhcpLinks[link.Attrs().Name] = match.DHCPArgs
		}
	}

	//run dhcp
	wg := sync.WaitGroup{}
	for iface, args := range dhcpLinks {
		wg.Add(1)
		go func(iface, args string) {
			runDhcp(netCfg, iface, args, setHostname, setDns)
			wg.Done()
		}(iface, args)
	}
	wg.Wait()

	return err
}

func runDhcp(netCfg *NetworkConfig, iface string, argstr string, setHostname, setDns bool) {
	log.Infof("Running DHCP on %s", iface)
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

	if !setDns {
		args = append(args, "--nohook", "resolv.conf")
	}

	args = append(args, iface)
	cmd := exec.Command(args[0], args[1:]...)
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

func setGateway(gateway string) error {
	if gateway == "" {
		return nil
	}

	gatewayIp := net.ParseIP(gateway)
	if gatewayIp == nil {
		return errors.New("Invalid gateway address " + gateway)
	}

	route := netlink.Route{
		Scope: netlink.SCOPE_UNIVERSE,
		Gw:    gatewayIp,
	}

	if err := netlink.RouteAdd(&route); err == syscall.EEXIST {
		//Ignore this error
	} else if err != nil {
		log.Errorf("gateway set failed: %v", err)
		return err
	}

	log.Infof("Set default gateway %s", gateway)
	return nil
}

func applyInterfaceConfig(link netlink.Link, netConf InterfaceConfig) error {
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

	if netConf.Bridge != "" && netConf.Bridge != "true" {
		b, err := NewBridge(netConf.Bridge)
		if err != nil {
			return err
		}
		if err := b.AddLink(link); err != nil {
			return err
		}
		return nil
	}

	if netConf.IPV4LL {
		if err := AssignLinkLocalIP(link); err != nil {
			log.Errorf("IPV4LL set failed: %v", err)
			return err
		}
	} else {
		addresses := []string{}

		if netConf.Address != "" {
			addresses = append(addresses, netConf.Address)
		}

		if len(netConf.Addresses) > 0 {
			addresses = append(addresses, netConf.Addresses...)
		}

		for _, address := range addresses {
			err := applyAddress(address, link, netConf)
			if err != nil {
				log.Errorf("Failed to apply address %s to %s: %v", address, link.Attrs().Name, err)
			}
		}
	}

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

	if err := setGateway(netConf.Gateway); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.Gateway)
	}

	if err := setGateway(netConf.GatewayIpv6); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.GatewayIpv6)
	}

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
			log.Errorf("Failed to run command [%s]: %v", cmd, err)
			continue
		}
	}
}
