package cloudinit

import (
	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/compose"
	"github.com/rancher/os/pkg/init/docker"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/sysinit"
	"github.com/rancher/os/pkg/util"
)

func CloudInit(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	stateConfig := config.LoadConfigWithPrefix(config.StateDir)
	cfg.Rancher.CloudInit.Datasources = stateConfig.Rancher.CloudInit.Datasources

	if stateConfig.Rancher.Network.DHCPTimeout > 0 {
		cfg.Rancher.Network.DHCPTimeout = stateConfig.Rancher.Network.DHCPTimeout
		if err := config.Set("rancher.network.dhcp_timeout", stateConfig.Rancher.Network.DHCPTimeout); err != nil {
			log.Error(err)
		}
	}

	if len(stateConfig.Rancher.Network.Interfaces) > 0 {
		cfg.Rancher.Network = stateConfig.Rancher.Network
		if err := config.Set("rancher.network", stateConfig.Rancher.Network); err != nil {
			log.Error(err)
		}
	}

	hypervisor := util.GetHypervisor()
	if hypervisor == "" {
		log.Infof("ros init: No Detected Hypervisor")
	} else {
		log.Infof("ros init: Detected Hypervisor: %s", hypervisor)
	}
	if hypervisor == "vmware" {
		// add vmware to the end - we don't want to over-ride an choices the user has made
		cfg.Rancher.CloudInit.Datasources = append(cfg.Rancher.CloudInit.Datasources, hypervisor)
	}

	if err := config.Set("rancher.cloud_init.datasources", cfg.Rancher.CloudInit.Datasources); err != nil {
		log.Error(err)
	}

	log.Infof("init, runCloudInitServices(%v)", cfg.Rancher.CloudInit.Datasources)
	if err := runCloudInitServices(cfg); err != nil {
		log.Error(err)
	}

	// It'd be nice to push to rsyslog before this, but we don't have network
	log.AddRSyslogHook()

	return config.LoadConfig(), nil
}

func runCloudInitServices(cfg *config.CloudConfig) error {
	c, err := docker.Start(cfg)
	if err != nil {
		return err
	}

	defer docker.Stop(c)

	_, err = config.ChainCfgFuncs(cfg,
		[]config.CfgFuncData{
			{"cloudinit loadImages", sysinit.LoadBootstrapImages},
			{"cloudinit Services", runCloudInitServiceSet},
		})
	return err
}

func runCloudInitServiceSet(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	log.Info("Running cloud-init services")
	_, err := compose.RunServiceSet("cloud-init", cfg, cfg.Rancher.CloudInitServices)
	return cfg, err
}
