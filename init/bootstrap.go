package init

import (
	"syscall"

	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
)

func bootstrapServices(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if util.ResolveDevice(cfg.Rancher.State.Dev) != "" && len(cfg.Bootcmd) == 0 {
		log.Info("NOT Running Bootstrap")

		return cfg, nil
	}
	log.Info("Running Bootstrap")
	_, err := compose.RunServiceSet("bootstrap", cfg, cfg.Rancher.BootstrapContainers)
	return cfg, err
}

func runCloudInitServiceSet(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	log.Info("Running cloud-init services")
	_, err := compose.RunServiceSet("cloud-init", cfg, cfg.Rancher.CloudInitServices)
	return cfg, err
}

func startDocker(cfg *config.CloudConfig) (chan interface{}, error) {
	launchConfig, args := getLaunchConfig(cfg, &cfg.Rancher.BootstrapDocker)
	launchConfig.Fork = true
	launchConfig.LogFile = ""
	launchConfig.NoLog = true

	cmd, err := dfs.LaunchDocker(launchConfig, config.SystemDockerBin, args...)
	if err != nil {
		return nil, err
	}

	c := make(chan interface{})
	go func() {
		<-c
		cmd.Process.Signal(syscall.SIGTERM)
		cmd.Wait()
		c <- struct{}{}
	}()

	return c, nil
}

func stopDocker(c chan interface{}) error {
	c <- struct{}{}
	<-c

	return nil
}

func bootstrap(cfg *config.CloudConfig) error {
	log.Info("Launching Bootstrap Docker")
	log.Infof("bootstrap container: Autoformat(%v)", cfg.Rancher.State.Autoformat)

	c, err := startDocker(cfg)
	if err != nil {
		return err
	}

	defer stopDocker(c)

	_, err = config.ChainCfgFuncs(cfg,
		map[string]config.CfgFunc{
			"bootstrap loadImages": loadImages,
			"bootstrap Services":   bootstrapServices,
		})
	return err
}

func runCloudInitServices(cfg *config.CloudConfig) error {
	c, err := startDocker(cfg)
	if err != nil {
		return err
	}

	defer stopDocker(c)

	_, err = config.ChainCfgFuncs(cfg,
		map[string]config.CfgFunc{
			"cloudinit loadImages": loadImages,
			"cloudinit Services":   runCloudInitServiceSet,
		})
	return err
}
