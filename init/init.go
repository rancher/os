// +build linux

package init

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/docker-from-scratch"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	STATE string = "/state"
)

var (
	mountConfig = dockerlaunch.Config{
		CgroupHierarchy: map[string]string{
			"cpu":      "cpu",
			"cpuacct":  "cpu",
			"net_cls":  "net_cls",
			"net_prio": "net_cls",
		},
	}
)

func loadModules(cfg *config.Config) error {
	mounted := map[string]bool{}

	f, err := os.Open("/proc/modules")
	if err != nil {
		return err
	}
	defer f.Close()

	reader := bufio.NewScanner(f)
	for reader.Scan() {
		mounted[strings.SplitN(reader.Text(), " ", 2)[0]] = true
	}

	for _, module := range cfg.Modules {
		if mounted[module] {
			continue
		}

		log.Debugf("Loading module %s", module)
		if err := exec.Command("modprobe", module).Run(); err != nil {
			log.Errorf("Could not load module %s, err %v", module, err)
		}
	}

	return nil
}

func sysInit(cfg *config.Config) error {
	args := append([]string{config.SYSINIT_BIN}, os.Args[1:]...)

	cmd := &exec.Cmd{
		Path: config.ROS_BIN,
		Args: args,
	}

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}

	return os.Stdin.Close()
}

func MainInit() {
	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func mountState(cfg *config.Config) error {
	var err error

	if cfg.State.Dev == "" {
		return nil
	}

	dev := util.ResolveDevice(cfg.State.Dev)
	if dev == "" {
		return fmt.Errorf("Could not resolve device %q", cfg.State.Dev)
	}
	fsType := cfg.State.FsType
	if fsType == "auto" {
		fsType, err = util.GetFsType(dev)
	}

	if err != nil {
		return err
	}

	log.Debugf("FsType has been set to %s", fsType)
	log.Infof("Mounting state device %s to %s", dev, STATE)
	return util.Mount(dev, STATE, fsType, "")
}

func tryMountState(cfg *config.Config) error {
	if mountState(cfg) == nil {
		return nil
	}

	// If we failed to mount lets run bootstrap and try again
	if err := bootstrap(cfg); err != nil {
		return err
	}

	return mountState(cfg)
}

func tryMountAndBootstrap(cfg *config.Config) error {
	if err := tryMountState(cfg); !cfg.State.Required && err != nil {
		return nil
	} else if err != nil {
		return err
	}

	log.Debugf("Switching to new root at %s", STATE)
	return switchRoot(STATE)
}

func getLaunchConfig(cfg *config.Config, dockerCfg *config.DockerConfig) (*dockerlaunch.Config, []string) {
	var launchConfig dockerlaunch.Config

	args := dockerlaunch.ParseConfig(&launchConfig, append(dockerCfg.Args, dockerCfg.ExtraArgs...)...)

	launchConfig.DnsConfig.Nameservers = cfg.Network.Dns.Nameservers
	launchConfig.DnsConfig.Search = cfg.Network.Dns.Search

	if !cfg.Debug {
		launchConfig.LogFile = config.SYSTEM_DOCKER_LOG
	}

	return &launchConfig, args
}

func RunInit() error {
	var cfg config.Config

	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")
	// Magic setting to tell Docker to do switch_root and not pivot_root
	os.Setenv("DOCKER_RAMDISK", "true")

	initFuncs := []config.InitFunc{
		func(cfg *config.Config) error {
			return dockerlaunch.PrepareFs(&mountConfig)
		},
		func(cfg *config.Config) error {
			newCfg, err := config.LoadConfig()
			if err == nil {
				newCfg, err = config.LoadConfig()
			}
			if err == nil {
				*cfg = *newCfg
			}

			if cfg.Debug {
				cfgString, _ := config.Dump(false, true)
				if cfgString != "" {
					log.Debugf("Config: %s", cfgString)
				}
			}

			return err
		},
		loadModules,
		tryMountAndBootstrap,
		func(cfg *config.Config) error {
			return cfg.Reload()
		},
		loadModules,
		sysInit,
	}

	if err := config.RunInitFuncs(&cfg, initFuncs); err != nil {
		return err
	}

	launchConfig, args := getLaunchConfig(&cfg, &cfg.SystemDocker)

	log.Info("Launching System Docker")
	_, err := dockerlaunch.LaunchDocker(launchConfig, config.DOCKER_BIN, args...)
	return err
}
