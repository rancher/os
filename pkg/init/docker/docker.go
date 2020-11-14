package docker

import (
	"syscall"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/dfs"
)

func Start(cfg *config.CloudConfig) (chan interface{}, error) {
	launchConfig, args := GetLaunchConfig(cfg, &cfg.Rancher.BootstrapDocker)
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

func Stop(c chan interface{}) error {
	c <- struct{}{}
	<-c

	return nil
}

func GetLaunchConfig(cfg *config.CloudConfig, dockerCfg *config.DockerConfig) (*dfs.Config, []string) {
	var launchConfig dfs.Config

	args := dfs.ParseConfig(&launchConfig, dockerCfg.FullArgs()...)

	launchConfig.DNSConfig.Nameservers = cfg.Rancher.Defaults.Network.DNS.Nameservers
	launchConfig.DNSConfig.Search = cfg.Rancher.Defaults.Network.DNS.Search
	launchConfig.Environment = dockerCfg.Environment

	if !cfg.Rancher.Debug {
		launchConfig.LogFile = cfg.Rancher.Defaults.SystemDockerLogs
	}

	return &launchConfig, args
}
