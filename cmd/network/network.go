package network

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/rancher/netconf"
	"github.com/rancherio/os/config"
)

func Main() {
	args := os.Args
	if len(args) > 1 {
		fmt.Println("call " + args[0] + " to load network config from rancher.yml config file")
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if err := netconf.ApplyNetworkConfigs(&cfg.Network); err != nil {
		log.Fatal(err)
	}
}
