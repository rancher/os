package network

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
	"github.com/ryanuber/go-glob"
	"github.com/vishvananda/netlink"
)

func Main() {
	args := os.Args
	if len(args) > 1 {
		fmt.Println("call " + args[0] + "to load network config from rancher.yml config file")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	applyNetworkConfigs(cfg)
}

func applyNetworkConfigs(cfg *config.Config) error {
	links, err := netlink.LinkList()
	if err != nil {
		return err
	}

	//apply network config
	for _, netConf := range cfg.Network.Interfaces {
		for _, link := range links {
			err := applyNetConf(link, netConf)
			if err != nil {
				log.Errorf("Failed to apply settings to %s : %v", link.Attrs().Name, err)
			}
		}
	}

	if err != nil {
		return err
	}

	//post run
	if cfg.Network.PostRun != nil {
		return docker.StartAndWait(config.DOCKER_HOST, cfg.Network.PostRun)
	}
	return nil
}

func applyNetConf(link netlink.Link, netConf config.InterfaceConfig) error {
	if matches(link.Attrs().Name, netConf.Match) {
		if netConf.DHCP {
			log.Infof("Running DHCP on %s", link.Attrs().Name)
			cmd := exec.Command("udhcpc", "-i", link.Attrs().Name, "-t", "20", "-n")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Error(err)
			}
		} else {
			if netConf.Address == "" {
				return errors.New("DHCP is false and Address is not set")
			}
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
	}
	return nil
}

func matches(link, conf string) bool {
	return glob.Glob(conf, link)
}
