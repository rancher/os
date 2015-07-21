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

func autoformat(cfg *config.Config) error {
	if len(cfg.State.Autoformat) == 0 || util.ResolveDevice(cfg.State.Dev) != "" {
		return nil
	}
	AUTOFORMAT := "AUTOFORMAT=" + strings.Join(cfg.State.Autoformat, " ")
	FORMATZERO := "FORMATZERO=" + fmt.Sprint(cfg.State.FormatZero)
	cfg.Autoformat["autoformat"].Environment = project.NewMaporEqualSlice([]string{AUTOFORMAT, FORMATZERO})
	err := docker.RunServices("autoformat", cfg, cfg.Autoformat)
	return err
}

func runBootstrapContainers(cfg *config.Config) error {
	return docker.RunServices("bootstrap", cfg, cfg.BootstrapContainers)
}

func startDocker(cfg *config.Config) (chan interface{}, error) {
	for _, d := range []string{config.DOCKER_SYSTEM_HOST, "/var/run"} {
		err := os.MkdirAll(d, 0700)
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.Command(cfg.BootstrapDocker.Args[0], cfg.BootstrapDocker.Args[1:]...)
	if cfg.Debug {
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

func bootstrap(cfg *config.Config) error {
	log.Info("Starting bootstrap")
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
