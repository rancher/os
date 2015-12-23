package netconf

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/flynn/go-shlex"

	"github.com/ryanuber/go-glob"
	"github.com/vishvananda/netlink"
)

const (
	CONF       = "/var/lib/rancher/conf"
	NET_SCRIPT = "/var/lib/rancher/conf/network.sh"
)

func createInterfaces(netCfg *NetworkConfig) error {
	for name, iface := range netCfg.Interfaces {
		if iface.Bridge {
			bridge := netlink.Bridge{}
			bridge.LinkAttrs.Name = name

			if err := netlink.LinkAdd(&bridge); err != nil {
				log.Errorf("Failed to create bridge %s: %v", name, err)
			}
		} else if iface.Bond != "" {
			bondIface, ok := netCfg.Interfaces[iface.Bond]
			if !ok {
				log.Errorf("Failed to find bond configuration for [%s]", iface.Bond)
				continue
			}
			bond := Bond(iface.Bond)
			if bond.Error() != nil {
				log.Errorf("Failed to create bond [%s]: %v", iface.Bond, bond.Error())
				continue
			}

			for k, v := range bondIface.BondOpts {
				bond.Opt(k, v)
				bond.Clear()
			}
		}
	}

	return nil
}

func runScript(netCfg *NetworkConfig) error {
	if netCfg.Script == "" {
		return nil
	}

	if _, err := os.Stat(CONF); os.IsNotExist(err) {
		if err := os.MkdirAll(CONF, 0700); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(NET_SCRIPT, []byte(netCfg.Script), 0700); err != nil {
		return err
	}

	cmd := exec.Command(NET_SCRIPT)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ApplyNetworkConfigs(netCfg *NetworkConfig) error {
	log.Debugf("Config: %#v", netCfg)
	if err := runScript(netCfg); err != nil {
		log.Errorf("Failed to run script: %v", err)
	}

	if err := createInterfaces(netCfg); err != nil {
		return err
	}

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	dhcpLinks := []string{}

	//apply network config
	for _, link := range links {
		linkName := link.Attrs().Name
		var match InterfaceConfig

		for key, netConf := range netCfg.Interfaces {
			if netConf.Match == "" {
				netConf.Match = key
			}

			if netConf.Match == "" {
				continue
			}

			if len(netConf.Match) > 4 && strings.ToLower(netConf.Match[:3]) == "mac" {
				haAddr, err := net.ParseMAC(netConf.Match[4:])
				if err != nil {
					return err
				}
				// Don't match mac address of the bond because it is the same as the slave
				if bytes.Compare(haAddr, link.Attrs().HardwareAddr) == 0 && link.Attrs().Name != netConf.Bond {
					// MAC address match is used over all other matches
					match = netConf
					break
				}
			}

			// "" means match has not been found
			if match.Match == "" && matches(linkName, netConf.Match) {
				match = netConf
			}

			if netConf.Match == linkName {
				// Found exact match, use it over wildcard match
				match = netConf
			}
		}

		if match.Match != "" {
			if match.DHCP {
				dhcpLinks = append(dhcpLinks, link.Attrs().Name)
			} else if err = applyNetConf(link, match); err != nil {
				log.Errorf("Failed to apply settings to %s : %v", linkName, err)
			}
		}
	}

	if len(dhcpLinks) > 0 {
		log.Infof("Running DHCP on %v", dhcpLinks)
		dhcpcdArgs := append([]string{"-MA4", "-e", "force_hostname=true"}, dhcpLinks...)
		cmd := exec.Command("dhcpcd", dhcpcdArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Error(err)
		}
	}

	if err != nil {
		return err
	}

	return nil
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

func applyNetConf(link netlink.Link, netConf InterfaceConfig) error {
	if netConf.Bond != "" {
		b := Bond(netConf.Bond)
		b.AddSlave(link.Attrs().Name)
		if b.Error() != nil {
			return b.Error()
		}

		return linkUp(link, netConf)
	}

	if netConf.IPV4LL {
		if err := AssignLinkLocalIP(link); err != nil {
			log.Errorf("IPV4LL set failed: %v", err)
			return err
		}
	} else if netConf.Address == "" && len(netConf.Addresses) == 0 {
		return nil
	} else {
		if netConf.Address != "" {
			err := applyAddress(netConf.Address, link, netConf)
			if err != nil {
				log.Errorf("Failed to apply address %s to %s: %v", netConf.Address, link.Attrs().Name, err)
			}
		}
		for _, address := range netConf.Addresses {
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

	if err := linkUp(link, netConf); err != nil {
		return err
	}

	if err := setGateway(netConf.Gateway); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.Gateway)
	}

	if err := setGateway(netConf.GatewayIpv6); err != nil {
		log.Errorf("Fail to set gateway %s", netConf.Gateway)
	}

	for _, postUp := range netConf.PostUp {
		postUp = strings.TrimSpace(postUp)
		if postUp == "" {
			continue
		}

		args, err := shlex.Split(strings.Replace(postUp, "$iface", link.Attrs().Name, -1))
		if err != nil {
			log.Errorf("Failed to parse command [%s]: %v", postUp, err)
			continue
		}

		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Failed to run command [%s]: %v", postUp, err)
			continue
		}
	}

	return nil
}

func matches(link, conf string) bool {
	return glob.Glob(conf, link)
}
