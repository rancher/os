package network

import (
	"golang.org/x/net/context"

	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
	"github.com/rancher/os/netconf"
)

func Main() {
	log.InitLogger()

	cfg := config.LoadConfig()
	ApplyNetworkConfig(cfg)

	log.Infof("Restart syslog")
	client, err := docker.NewSystemClient()
	if err != nil {
		log.Error(err)
	}

	if err := client.ContainerRestart(context.Background(), "syslog", 10); err != nil {
		log.Error(err)
	}

	select {}
}

func ApplyNetworkConfig(cfg *config.CloudConfig) {
	log.Infof("Apply Network Config")
	nameservers := cfg.Rancher.Network.DNS.Nameservers
	search := cfg.Rancher.Network.DNS.Search
	userSetDNS := len(nameservers) > 0 || len(search) > 0
	if !userSetDNS {
		nameservers = cfg.Rancher.Defaults.Network.DNS.Nameservers
		search = cfg.Rancher.Defaults.Network.DNS.Search
	}

	// TODO: don't write to the file if nameservers is still empty
	if _, err := resolvconf.Build("/etc/resolv.conf", nameservers, search, nil); err != nil {
		log.Error(err)
	}

	if err := hostname.SetHostnameFromCloudConfig(cfg); err != nil {
		log.Error(err)
	}

	if err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network); err != nil {
		log.Error(err)
	}

	// TODO: seems wrong to do this outside netconf
	userSetHostname := cfg.Hostname != ""
	log.Infof("Apply Network Config RunDhcp")
	if err := netconf.RunDhcp(&cfg.Rancher.Network, !userSetHostname, !userSetDNS); err != nil {
		log.Error(err)
	}

	log.Infof("Apply Network Config SyncHostname")
	if err := hostname.SyncHostname(); err != nil {
		log.Error(err)
	}
}
