package network

import (
	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/netconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
)

func Main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	nameservers := cfg.Rancher.Network.Dns.Nameservers
	search := cfg.Rancher.Network.Dns.Search
	if len(nameservers) == 0 && len(search) == 0 {
		nameservers = cfg.Rancher.DefaultNetwork.Dns.Nameservers
		search = cfg.Rancher.DefaultNetwork.Dns.Search
	}

	if _, err := resolvconf.Build("/etc/resolv.conf", nameservers, search, nil); err != nil {
		log.Error(err)
	}

	if err := hostname.SetHostnameFromCloudConfig(cfg); err != nil {
		log.Error(err)
	}

	if err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network); err != nil {
		log.Error(err)
	}
}
