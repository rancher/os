package init

import (
	"os"
	"os/exec"
	"syscall"

	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/librcompose/project"
	"strings"
)

func autoformat(cfg *config.CloudConfig) error {
	if len(cfg.Rancher.State.Autoformat) == 0 || util.ResolveDevice(cfg.Rancher.State.Dev) != "" {
		return nil
	}
	AUTOFORMAT := "AUTOFORMAT=" + strings.Join(cfg.Rancher.State.Autoformat, " ")
	FORMATZERO := "FORMATZERO=" + fmt.Sprint(cfg.Rancher.State.FormatZero)
	cfg.Rancher.Autoformat["autoformat"].Environment = project.NewMaporEqualSlice([]string{AUTOFORMAT, FORMATZERO})
	log.Info("Running Autoformat services")
	err := docker.RunServices("autoformat", cfg, cfg.Rancher.Autoformat)
	return err
}

func runBootstrapContainers(cfg *config.CloudConfig) error {
	log.Info("Running Bootstrap services")
	return docker.RunServices("bootstrap", cfg, cfg.Rancher.BootstrapContainers)
}

func startDocker(cfg *config.CloudConfig) (chan interface{}, error) {
	for _, d := range []string{config.DOCKER_SYSTEM_HOST, "/var/run"} {
		err := os.MkdirAll(d, 0700)
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.Command(cfg.Rancher.BootstrapDocker.Args[0], cfg.Rancher.BootstrapDocker.Args[1:]...)
	if cfg.Rancher.Debug {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
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

	return os.RemoveAll(config.DOCKER_SYSTEM_HOME)
}

func bootstrap(cfg *config.CloudConfig) error {
	log.Info("Launching Bootstrap Docker")
	c, err := startDocker(cfg)
	if err != nil {
		return err
	}

	initFuncs := []config.InitFunc{
		loadImages,
		runBootstrapContainers,
		autoformat,
	}

	defer stopDocker(c)

	return config.RunInitFuncs(cfg, initFuncs)
}
