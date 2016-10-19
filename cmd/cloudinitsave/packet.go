package cloudinitsave

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/Sirupsen/logrus"
	"github.com/packethost/packngo/metadata"
	"github.com/rancher/os/config"
	"github.com/rancher/os/netconf"
)

func enablePacketNetwork(cfg *config.RancherConfig) {
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

	bondCfg := config.InterfaceConfig{
		Addresses: []string{},
		BondOpts: map[string]string{
			"lacp_rate":        "1",
			"xmit_hash_policy": "layer3+4",
			"downdelay":        "200",
			"updelay":          "200",
			"miimon":           "100",
			"mode":             "4",
		},
	}
	netCfg := config.NetworkConfig{
		Interfaces: map[string]config.InterfaceConfig{},
	}
	for _, iface := range m.Network.Interfaces {
		netCfg.Interfaces["mac="+iface.Mac] = config.InterfaceConfig{
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
	b, _ := yaml.Marshal(netCfg)
	logrus.Debugf("Generated network config: %s", string(b))

	cc := config.CloudConfig{
		Rancher: config.RancherConfig{
			Network: netCfg,
		},
	}

	// Post to phone home URL on first boot
	if _, err = os.Stat(config.CloudConfigNetworkFile); err != nil {
		if _, err = http.Post(m.PhoneHomeURL, "application/json", bytes.NewReader([]byte{})); err != nil {
			logrus.Errorf("Failed to post to Packet phone home URL: %v", err)
		}
	}

	if err := os.MkdirAll(path.Dir(config.CloudConfigNetworkFile), 0700); err != nil {
		logrus.Errorf("Failed to create directory for file %s: %v", config.CloudConfigNetworkFile, err)
	}

	if err := config.WriteToFile(cc, config.CloudConfigNetworkFile); err != nil {
		logrus.Errorf("Failed to save config file %s: %v", config.CloudConfigNetworkFile, err)
	}
}
