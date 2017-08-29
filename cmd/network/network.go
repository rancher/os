package network

import (
	"golang.org/x/net/context"

	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
	"github.com/rancher/os/netconf"
	"io/ioutil"
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
	userSetDNS := len(cfg.Rancher.Network.DNS.Nameservers) > 0 || len(cfg.Rancher.Network.DNS.Search) > 0
	if userSetDNS {
		if _, err := resolvconf.Build("/etc/resolv.conf", cfg.Rancher.Network.DNS.Nameservers, cfg.Rancher.Network.DNS.Search, nil); err != nil {
			log.Error(err)
		}
	}

	if err := hostname.SetHostnameFromCloudConfig(cfg); err != nil {
		log.Error(err)
	}

	userSetHostname := cfg.Hostname != ""
	dhcpSetDNS, err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network, userSetHostname, userSetDNS)
	if err != nil {
		log.Error(err)
	}

	if dhcpSetDNS {
		log.Infof("DNS set by DHCP")
	}

	if !userSetDNS && !dhcpSetDNS {
		// only write 8.8.8.8,8.8.4.4 as a last resort
		log.Infof("Writing default resolv.conf - no user setting, and no DHCP setting")
		if _, err := resolvconf.Build("/etc/resolv.conf",
			cfg.Rancher.Defaults.Network.DNS.Nameservers,
			cfg.Rancher.Defaults.Network.DNS.Search,
			nil); err != nil {
			log.Error(err)
		}
	}
	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Debugf("Resolve.conf == [%s], %s", resolve, err)

	log.Infof("Apply Network Config SyncHostname")
	if err := hostname.SyncHostname(); err != nil {
		log.Error(err)
	}
}
