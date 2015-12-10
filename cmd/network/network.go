package network

import (
	"fmt"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/netconf"
	"github.com/rancher/os/cmd/cloudinit"
	"github.com/rancher/os/config"
)

const (
	NETWORK_DONE     = "/var/run/network.done"
	WAIT_FOR_NETWORK = "wait-for-network"
)

func sendTerm(proc string) {
	cmd := exec.Command("killall", "-TERM", proc)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func Main() {
	args := os.Args
	if len(args) > 1 {
		fmt.Println("call " + args[0] + " to load network config from cloud-config.yml")
		return
	}
	os.Remove(NETWORK_DONE) // ignore error
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	cloudinit.SetHostname(cfg) // ignore error
	if err := netconf.ApplyNetworkConfigs(&cfg.Rancher.Network); err != nil {
		log.Error(err)
	}
	if cfg.Rancher.Network.Dns.Override {
		log.WithFields(log.Fields{"nameservers": cfg.Rancher.Network.Dns.Nameservers}).Info("Override nameservers")
		if _, err := resolvconf.Build("/etc/resolv.conf", cfg.Rancher.Network.Dns.Nameservers, cfg.Rancher.Network.Dns.Search, nil); err != nil {
			log.Error(err)
		}
	}
	if _, err := os.Create(NETWORK_DONE); err != nil {
		log.Error(err)
	}
	sendTerm(WAIT_FOR_NETWORK)
	select {}
}
