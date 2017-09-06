package init

import (
//	"syscall"

//	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
//	"github.com/rancher/os/dfs"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
)

func bootstrap(cfg *config.CloudConfig) error {
	log.Info("Launching Bootstrap Docker")

	if util.ResolveDevice(cfg.Rancher.State.Dev) != "" && len(cfg.Bootcmd) == 0 {
		log.Info("NOT Running Bootstrap")

		return nil
	}
	log.Info("Running Bootstrap")
//	_, err := runServiceSet("bootstrap", cfg, cfg.Rancher.BootstrapContainers)
//	return err
	return nil
}

func runCloudInitServices(cfg *config.CloudConfig) error {
	log.Info("Running cloud-init services")
//	_, err := runServiceSet("cloud-init", cfg, cfg.Rancher.CloudInitServices)

//	return err
	return nil
}

