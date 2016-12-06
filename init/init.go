// +build linux

package init

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/docker-from-scratch"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

const (
	STATE string = "/state"

	TMPFS_MAGIC int64 = 0x01021994
	RAMFS_MAGIC int64 = 0x858458f6
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

func loadModules(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	mounted := map[string]bool{}

	f, err := os.Open("/proc/modules")
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	reader := bufio.NewScanner(f)
	for reader.Scan() {
		mounted[strings.SplitN(reader.Text(), " ", 2)[0]] = true
	}

	for _, module := range cfg.Rancher.Modules {
		if mounted[module] {
			continue
		}

		log.Debugf("Loading module %s", module)
		if err := exec.Command("modprobe", module).Run(); err != nil {
			log.Errorf("Could not load module %s, err %v", module, err)
		}
	}

	return cfg, nil
}

func sysInit(c *config.CloudConfig) (*config.CloudConfig, error) {
	args := append([]string{config.SYSINIT_BIN}, os.Args[1:]...)

	cmd := &exec.Cmd{
		Path: config.ROS_BIN,
		Args: args,
	}

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return c, err
	}

	return c, os.Stdin.Close()
}

func MainInit() {
	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func mountConfigured(display, dev, fsType, target string) error {
	var err error

	if dev == "" {
		return nil
	}

	dev = util.ResolveDevice(dev)
	if dev == "" {
		return fmt.Errorf("Could not resolve device %q", dev)
	}
	if fsType == "auto" {
		fsType, err = util.GetFsType(dev)
	}

	if err != nil {
		return err
	}

	log.Debugf("FsType has been set to %s", fsType)
	log.Infof("Mounting %s device %s to %s", display, dev, target)
	return util.Mount(dev, target, fsType, "")
}

func mountState(cfg *config.CloudConfig) error {
	return mountConfigured("state", cfg.Rancher.State.Dev, cfg.Rancher.State.FsType, STATE)
}

func mountOem(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if err := mountConfigured("oem", cfg.Rancher.State.OemDev, cfg.Rancher.State.OemFsType, config.OEM); err != nil {
		log.Debugf("Not mounting OEM: %v", err)
	} else {
		log.Infof("Mounted OEM: %s", cfg.Rancher.State.OemDev)
	}
	return cfg, nil
}

func tryMountState(cfg *config.CloudConfig) error {
	if mountState(cfg) == nil {
		return nil
	}

	// If we failed to mount lets run bootstrap and try again
	if err := bootstrap(cfg); err != nil {
		return err
	}

	return mountState(cfg)
}

func tryMountAndBootstrap(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if isInitrd() {
		if err := tryMountState(cfg); !cfg.Rancher.State.Required && err != nil {
			return cfg, nil
		} else if err != nil {
			return cfg, err
		}

		log.Debugf("Switching to new root at %s %s", STATE, cfg.Rancher.State.Directory)
		if err := switchRoot(STATE, cfg.Rancher.State.Directory, cfg.Rancher.RmUsr); err != nil {
			return cfg, err
		}
	}
	return mountOem(cfg)
}

func getLaunchConfig(cfg *config.CloudConfig, dockerCfg *config.DockerConfig) (*dockerlaunch.Config, []string) {
	var launchConfig dockerlaunch.Config

	args := dockerlaunch.ParseConfig(&launchConfig, append(dockerCfg.Args, dockerCfg.ExtraArgs...)...)

	launchConfig.DnsConfig.Nameservers = cfg.Rancher.Defaults.Network.Dns.Nameservers
	launchConfig.DnsConfig.Search = cfg.Rancher.Defaults.Network.Dns.Search
	launchConfig.Environment = dockerCfg.Environment
	launchConfig.EmulateSystemd = true

	if !cfg.Rancher.Debug {
		launchConfig.LogFile = config.SYSTEM_DOCKER_LOG
	}

	return &launchConfig, args
}

func isInitrd() bool {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	return int64(stat.Type) == TMPFS_MAGIC || int64(stat.Type) == RAMFS_MAGIC
}

func RunInit() error {
	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")
	if isInitrd() {
		log.Debug("Booting off an in-memory filesystem")
		// Magic setting to tell Docker to do switch_root and not pivot_root
		os.Setenv("DOCKER_RAMDISK", "true")
	} else {
		log.Debug("Booting off a persistent filesystem")
	}

	initFuncs := []config.CfgFunc{
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dockerlaunch.PrepareFs(&mountConfig)
		},
		mountOem,
		func(_ *config.CloudConfig) (*config.CloudConfig, error) {
			cfg := config.LoadConfig()

			if cfg.Rancher.Debug {
				cfgString, err := config.Export(false, true)
				if err != nil {
					log.WithFields(log.Fields{"err": err}).Error("Error serializing config")
				} else {
					log.Debugf("Config: %s", cfgString)
				}
			}

			return cfg, nil
		},
		loadModules,
		tryMountAndBootstrap,
		func(_ *config.CloudConfig) (*config.CloudConfig, error) {
			return config.LoadConfig(), nil
		},
		loadModules,
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dockerlaunch.PrepareFs(&mountConfig)
		},
		initializeSelinux,
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, syscall.Mount("", "/", "", syscall.MS_SHARED|syscall.MS_REC, "")
		},
		sysInit,
	}

	cfg, err := config.ChainCfgFuncs(nil, initFuncs...)
	if err != nil {
		return err
	}

	launchConfig, args := getLaunchConfig(cfg, &cfg.Rancher.SystemDocker)
	launchConfig.Fork = !cfg.Rancher.SystemDocker.Exec

	log.Info("Launching System Docker")
	_, err = dockerlaunch.LaunchDocker(launchConfig, config.SYSTEM_DOCKER_BIN, args...)
	if err != nil {
		return err
	}

	return pidOne()
}
