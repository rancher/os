package dhcp

import (
	"flag"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/rancher/netconf"
	"github.com/rancher/os/config"
	"github.com/rancher/os/hostname"
)

var (
	daemon bool
	flags  *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&daemon, "daemon", false, "run dhcpcd as daemon")
}

func Main() {
	flags.Parse(os.Args[1:])

	log.Infof("Running dhcp: daemon=%v", daemon)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	nameservers := cfg.Rancher.Network.Dns.Nameservers
	search := cfg.Rancher.Network.Dns.Search
	userSetDns := len(nameservers) > 0 || len(search) > 0
	userSetHostname := cfg.Hostname != ""

	if err := netconf.RunDhcp(&cfg.Rancher.Network, !userSetHostname, !userSetDns); err != nil {
		log.Error(err)
	}

	if err := hostname.SyncHostname(); err != nil {
		log.Error(err)
	}

	if daemon {
		select {}
	}
}
