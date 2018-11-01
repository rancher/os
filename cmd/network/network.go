package network

import (
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/docker"
	"github.com/rancher/os/pkg/hostname"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/netconf"

	"github.com/docker/libnetwork/resolvconf"
	"golang.org/x/net/context"
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

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM)
	<-signalChan
	log.Info("Received SIGTERM, shutting down")
	netconf.StopDhcpcd()
}

func ApplyNetworkConfig(cfg *config.CloudConfig) {
	log.Infof("Apply Network Config")
	userSetDNS := len(cfg.Rancher.Network.DNS.Nameservers) > 0 || len(cfg.Rancher.Network.DNS.Search) > 0

	if err := hostname.SetHostnameFromCloudConfig(cfg); err != nil {
		log.Errorf("Failed to set hostname from cloud config: %v", err)
	}

	userSetHostname := cfg.Hostname != ""
	if cfg.Rancher.Network.DHCPTimeout <= 0 {
		cfg.Rancher.Network.DHCPTimeout = cfg.Rancher.Defaults.Network.DHCPTimeout
	}
	dhcpSetDNS, err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network, userSetHostname, userSetDNS)
	if err != nil {
		log.Errorf("Failed to apply network configs(by netconf): %v", err)
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
			log.Errorf("Failed to write resolv.conf (!userSetDNS and !dhcpSetDNS): %v", err)
		}
	}
	if userSetDNS {
		if _, err := resolvconf.Build("/etc/resolv.conf", cfg.Rancher.Network.DNS.Nameservers, cfg.Rancher.Network.DNS.Search, nil); err != nil {
			log.Errorf("Failed to write resolv.conf (userSetDNS): %v", err)
		} else {
			log.Infof("writing to /etc/resolv.conf: nameservers: %v, search: %v", cfg.Rancher.Network.DNS.Nameservers, cfg.Rancher.Network.DNS.Search)
		}
	}

	resolve, err := ioutil.ReadFile("/etc/resolv.conf")
	log.Debugf("Resolve.conf == [%s], %v", resolve, err)

	log.Infof("Apply Network Config SyncHostname")
	if err := hostname.SyncHostname(); err != nil {
		log.Errorf("Failed to sync hostname: %v", err)
	}
}
