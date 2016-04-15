package network

import (
	"flag"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/netconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
)

const (
	NETWORK_DONE     = "/var/run/network.done"
	WAIT_FOR_NETWORK = "wait-for-network"
)

var (
	daemon bool
	flags  *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&daemon, "daemon", false, "run dhcpd as daemon")
}

func sendTerm(proc string) {
	cmd := exec.Command("killall", "-TERM", proc)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func Main() {
	flags.Parse(os.Args[1:])

	log.Infof("Running network: daemon=%v", daemon)

	os.Remove(NETWORK_DONE) // ignore error
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

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

	if f, err := os.Create(NETWORK_DONE); err != nil {
		log.Error(err)
	} else {
		f.Close()
	}
	sendTerm(WAIT_FOR_NETWORK)

	if daemon {
		select {}
	}
}
