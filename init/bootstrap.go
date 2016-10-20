package init

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/util"
)

func bootstrapServices(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if len(cfg.Rancher.State.Autoformat) == 0 || util.ResolveDevice(cfg.Rancher.State.Dev) != "" {
		return cfg, nil
	}
	log.Info("Running Bootstrap")
	_, err := compose.RunServiceSet("bootstrap", cfg, cfg.Rancher.BootstrapContainers)
	return cfg, err
}

func startDocker(cfg *config.CloudConfig) (chan interface{}, error) {
	launchConfig, args := getLaunchConfig(cfg, &cfg.Rancher.BootstrapDocker)
	launchConfig.Fork = true
	launchConfig.LogFile = ""
	launchConfig.NoLog = true

	cmd, err := dfs.LaunchDocker(launchConfig, config.SYSTEM_DOCKER_BIN, args...)
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
	c, err := startDocker(cfg)
	if err != nil {
		return err
	}

	defer stopDocker(c)

	_, err = config.ChainCfgFuncs(cfg,
		loadImages,
		bootstrapServices)
	return err
}
