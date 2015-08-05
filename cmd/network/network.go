package network

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
	"github.com/ryanuber/go-glob"
	"github.com/vishvananda/netlink"
)

func Main() {
	args := os.Args
	if len(args) > 1 {
		fmt.Println("call " + args[0] + " to load network config from cloud-config.yml")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	ApplyNetworkConfigs(&cfg.Rancher.Network)
}

func createInterfaces(netCfg *config.NetworkConfig) error {
	for name, iface := range netCfg.Interfaces {
		if !iface.Bridge {
			continue
		}

		bridge := netlink.Bridge{}
		bridge.LinkAttrs.Name = name

		if err := netlink.LinkAdd(&bridge); err != nil {
			log.Errorf("Failed to create bridge %s: %v", name, err)
		}
	}

	return nil
}

func ApplyNetworkConfigs(netCfg *config.NetworkConfig) error {
	if err := createInterfaces(netCfg); err != nil {
		return err
	}

	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	//apply network config
	for _, link := range links {
		linkName := link.Attrs().Name
		var match config.InterfaceConfig

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
				if bytes.Compare(haAddr, link.Attrs().HardwareAddr) == 0 {
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
			err = applyNetConf(link, match)
			if err != nil {
				log.Errorf("Failed to apply settings to %s : %v", linkName, err)
			}
		}
	}

	if err != nil {
		return err
	}

	//post run
	if netCfg.PostRun != nil {
		return docker.StartAndWait(config.DOCKER_SYSTEM_HOST, netCfg.PostRun)
	}
	return nil
}

func applyNetConf(link netlink.Link, netConf config.InterfaceConfig) error {
	if netConf.DHCP {
		log.Infof("Running DHCP on %s", link.Attrs().Name)
		cmd := exec.Command("dhcpcd", "-A4", "-e", "force_hostname=true", link.Attrs().Name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Error(err)
		}
	} else if netConf.IPV4LL {
		if err := AssignLinkLocalIP(link); err != nil {
			log.Error("IPV4LL set failed")
			return err
		}
	} else if netConf.Address == "" {
		return nil
	} else {
		addr, err := netlink.ParseAddr(netConf.Address)
		if err != nil {
			return err
		}
		if err := netlink.AddrAdd(link, addr); err != nil {
			log.Error("addr add failed")
			return err
		}
		log.Infof("Set %s on %s", netConf.Address, link.Attrs().Name)
	}

	if netConf.MTU > 0 {
		if err := netlink.LinkSetMTU(link, netConf.MTU); err != nil {
			log.Error("set MTU Failed")
			return err
		}
	}

	if err := netlink.LinkSetUp(link); err != nil {
		log.Error("failed to setup link")
		return err
	}

	if netConf.Gateway != "" {
		gatewayIp := net.ParseIP(netConf.Gateway)
		if gatewayIp == nil {
			return errors.New("Invalid gateway address " + netConf.Gateway)
		}

		route := netlink.Route{
			Scope: netlink.SCOPE_UNIVERSE,
			Gw:    net.ParseIP(netConf.Gateway),
		}
		if err := netlink.RouteAdd(&route); err != nil {
			log.Error("gateway set failed")
			return err
		}

		log.Infof("Set default gateway %s", netConf.Gateway)
	}

	return nil
}

func matches(link, conf string) bool {
	return glob.Glob(conf, link)
}
