package network

import (
	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
	"github.com/rancher/os/netconf"
)

func Main() {
	log.Infof("Running network")

	cfg := config.LoadConfig()
	ApplyNetworkConfig(cfg)

	select {}
}

func ApplyNetworkConfig(cfg *config.CloudConfig) {
	nameservers := cfg.Rancher.Network.Dns.Nameservers
	search := cfg.Rancher.Network.Dns.Search
	userSetDns := len(nameservers) > 0 || len(search) > 0
	if !userSetDns {
		nameservers = cfg.Rancher.Defaults.Network.Dns.Nameservers
		search = cfg.Rancher.Defaults.Network.Dns.Search
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

	userSetHostname := cfg.Hostname != ""
	if err := netconf.RunDhcp(&cfg.Rancher.Network, !userSetHostname, !userSetDns); err != nil {
		log.Error(err)
	}

	if err := hostname.SyncHostname(); err != nil {
		log.Error(err)
	}
}
