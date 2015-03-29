package init

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"github.com/rancherio/rancher-compose/project"
)

func autoformat(cfg *config.Config) error {
	if len(cfg.State.Autoformat) == 0 || util.ResolveDevice(cfg.State.Dev) != "" {
		return nil
	}

	var format string

outer:
	for _, dev := range cfg.State.Autoformat {
		log.Infof("Checking %s to auto-format", dev)
		if _, err := os.Stat(dev); os.IsNotExist(err) {
			continue
		}

		f, err := os.Open(dev)
		if err != nil {
			return err
		}
		defer f.Close()

		buffer := make([]byte, 1048576, 1048576)
		c, err := f.Read(buffer)
		if err != nil {
			return err
		}

		if c != 1048576 {
			log.Infof("%s not right size", dev)
			continue
		}

		for _, b := range buffer {
			if b != 0 {
				log.Infof("%s not empty", dev)
				continue outer
			}
		}

		format = dev
		break
	}

	if format != "" {
		log.Infof("Auto formatting : %s", format)
		return runServices("autoformat", cfg, map[string]*project.ServiceConfig{
			"autoformat": {
				Net:        "none",
				Privileged: true,
				Image:      "autoformat",
				Command:    format,
			},
			"udev": cfg.BootstrapContainers["udev"],
		})
	}

	return nil
}

func runBootstrapContainers(cfg *config.Config) error {
	return runServices("bootstrap", cfg, cfg.BootstrapContainers)
}

func startDocker(cfg *config.Config) (chan interface{}, error) {
	for _, d := range []string{config.DOCKER_SYSTEM_HOST, "/var/run"} {
		err := os.MkdirAll(d, 0700)
		if err != nil {
			return nil, err
		}
	}

	cmd := exec.Command(cfg.SystemDocker.Args[0], cfg.SystemDocker.Args[1:]...)
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
