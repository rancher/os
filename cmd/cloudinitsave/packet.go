package cloudinitsave

import (
	"github.com/rancher/os/log"

	"github.com/rancher/os/config"
	"github.com/rancher/os/netconf"
)

func enablePacketNetwork(cfg *config.RancherConfig) {
	bootStrapped := false
	for _, v := range cfg.Network.Interfaces {
		if v.Address != "" {
			if err := netconf.ApplyNetworkConfigs(&cfg.Network); err != nil {
				log.Errorf("Failed to bootstrap network: %v", err)
				return
			}
			bootStrapped = true
			break
		}
	}

	if !bootStrapped {
		return
	}

	// Post to phone home URL on first boot
	/*
		// TODO: bring this back
		if _, err = os.Stat(config.CloudConfigNetworkFile); err != nil {
			if _, err = http.Post(m.PhoneHomeURL, "application/json", bytes.NewReader([]byte{})); err != nil {
				log.Errorf("Failed to post to Packet phone home URL: %v", err)
			}
		}
	*/

	/*
		cc := config.CloudConfig{
			Rancher: config.RancherConfig{
				Network: netCfg,
			},
		}

		if err := os.MkdirAll(path.Dir(config.CloudConfigNetworkFile), 0700); err != nil {
			log.Errorf("Failed to create directory for file %s: %v", config.CloudConfigNetworkFile, err)
		}

		if err := config.WriteToFile(cc, config.CloudConfigNetworkFile); err != nil {
			log.Errorf("Failed to save config file %s: %v", config.CloudConfigNetworkFile, err)
		}
	*/
}
