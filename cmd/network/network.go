package network

import (
	"flag"
	"os"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/netconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/hostname"
)

var (
	stopNetworkPre bool
	flags          *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&stopNetworkPre, "stop-network-pre", false, "")
}

func Main() {
	flags.Parse(os.Args[1:])

	log.Infof("Running network: stop-network-pre=%v", stopNetworkPre)

	if stopNetworkPre {
		client, err := docker.NewSystemClient()
		if err != nil {
			log.Error(err)
		}

		err = client.ContainerStop(context.Background(), "network-pre", 10)
		if err != nil {
			log.Error(err)
		}

		_, err = client.ContainerWait(context.Background(), "network-pre")
		if err != nil {
			log.Error(err)
		}
	}

	cfg := config.LoadConfig()

	nameservers := cfg.Rancher.Network.Dns.Nameservers
	search := cfg.Rancher.Network.Dns.Search
	userSetDns := len(nameservers) > 0 || len(search) > 0
	if !userSetDns {
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

	userSetHostname := cfg.Hostname != ""
	if err := netconf.RunDhcp(&cfg.Rancher.Network, !userSetHostname, !userSetDns); err != nil {
		log.Error(err)
	}

	if err := hostname.SyncHostname(); err != nil {
		log.Error(err)
	}

	select {}
}
