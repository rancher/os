package cloudinit

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/Sirupsen/logrus"
	"github.com/packethost/packngo/metadata"
	"github.com/rancher/netconf"
	rancherConfig "github.com/rancher/os/config"
)

func enablePacketNetwork(cfg *rancherConfig.RancherConfig) {
	bootStrapped := false
	for _, v := range cfg.Network.Interfaces {
		if v.Address != "" {
			if err := netconf.ApplyNetworkConfigs(&cfg.Network); err != nil {
				logrus.Errorf("Failed to bootstrap network: %v", err)
				return
			}
			bootStrapped = true
			break
		}
	}

	if !bootStrapped {
		return
	}

	c := metadata.NewClient(http.DefaultClient)
	m, err := c.Metadata.Get()
	if err != nil {
		logrus.Errorf("Failed to get Packet metadata: %v", err)
		return
	}

	bondCfg := netconf.InterfaceConfig{
		Addresses: []string{},
		BondOpts: map[string]string{
			"lacp-rate":        "1",
			"xmit_hash_policy": "layer3+4",
			"downdelay":        "200",
			"updelay":          "200",
			"miimon":           "100",
			"mode":             "4",
		},
	}
	netCfg := netconf.NetworkConfig{
		Interfaces: map[string]netconf.InterfaceConfig{},
	}
	for _, iface := range m.Network.Interfaces {
		netCfg.Interfaces["mac="+iface.Mac] = netconf.InterfaceConfig{
			Bond: "bond0",
		}
	}
	for _, addr := range m.Network.Addresses {
		bondCfg.Addresses = append(bondCfg.Addresses, fmt.Sprintf("%s/%d", addr.Address, addr.Cidr))
		if addr.Gateway != "" {
			if addr.AddressFamily == 4 {
				if addr.Public {
					bondCfg.Gateway = addr.Gateway
				}
			} else {
				bondCfg.GatewayIpv6 = addr.Gateway
			}
		}

		if addr.AddressFamily == 4 && strings.HasPrefix(addr.Gateway, "10.") {
			bondCfg.PostUp = append(bondCfg.PostUp, "ip route add 10.0.0.0/8 via "+addr.Gateway)
		}
	}

	netCfg.Interfaces["bond0"] = bondCfg
	bytes, _ := yaml.Marshal(netCfg)
	logrus.Debugf("Generated network config: %s", string(bytes))

	cc := rancherConfig.CloudConfig{
		Rancher: rancherConfig.RancherConfig{
			Network: netCfg,
		},
	}

	if err := os.MkdirAll(path.Dir(rancherConfig.CloudConfigNetworkFile, 0700)); err != nil {
		logrus.Errorf("Failed to create directory for file %s: %v", rancherConfig.CloudConfigNetworkFile, err)
	}

	if err := rancherConfig.WriteToFile(cc, rancherConfig.CloudConfigNetworkFile); err != nil {
		logrus.Errorf("Failed to save config file %s: %v", rancherConfig.CloudConfigNetworkFile, err)
	}
}
