// +build linux

package init

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/mount"
	"github.com/rancher/os/cmd/cloudinitsave"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

const (
	STATE             string = "/state"
	BOOT2DOCKER_MAGIC string = "boot2docker, please format-me"

	TMPFS_MAGIC int64 = 0x01021994
	RAMFS_MAGIC int64 = 0x858458f6
)

var (
	mountConfig = dfs.Config{
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

func tryMountAndBootstrap(cfg *config.CloudConfig) (*config.CloudConfig, bool, error) {
	if !isInitrd() || cfg.Rancher.State.Dev == "" {
		return cfg, false, nil
	}

	if err := tryMountState(cfg); !cfg.Rancher.State.Required && err != nil {
		return cfg, false, nil
	} else if err != nil {
		return cfg, false, err
	}

	return cfg, true, nil
}

func getLaunchConfig(cfg *config.CloudConfig, dockerCfg *config.DockerConfig) (*dfs.Config, []string) {
	var launchConfig dfs.Config

	args := dfs.ParseConfig(&launchConfig, dockerCfg.FullArgs()...)

	launchConfig.DnsConfig.Nameservers = cfg.Rancher.Defaults.Network.Dns.Nameservers
	launchConfig.DnsConfig.Search = cfg.Rancher.Defaults.Network.Dns.Search
	launchConfig.Environment = dockerCfg.Environment

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

func setupSharedRoot(c *config.CloudConfig) (*config.CloudConfig, error) {
	if c.Rancher.NoSharedRoot {
		return c, nil
	}

	if isInitrd() {
		for _, i := range []string{"/mnt", "/media", "/var/lib/system-docker"} {
			if err := os.MkdirAll(i, 0755); err != nil {
				return c, err
			}
			if err := mount.Mount("tmpfs", i, "tmpfs", "rw"); err != nil {
				return c, err
			}
			if err := mount.MakeShared(i); err != nil {
				return c, err
			}
		}
		return c, nil
	}

	return c, mount.MakeShared("/")
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

	boot2DockerEnvironment := false
	var shouldSwitchRoot bool
	var cloudConfigBootFile []byte
	var metadataFile []byte
	initFuncs := []config.CfgFunc{
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dfs.PrepareFs(&mountConfig)
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
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if util.ResolveDevice("LABEL=B2D_STATE") != "" {
				boot2DockerEnvironment = true
				cfg.Rancher.State.Dev = "LABEL=B2D_STATE"
				return cfg, nil
			}

			devices := []string{"/dev/sda", "/dev/vda"}
			data := make([]byte, len(BOOT2DOCKER_MAGIC))

			for _, device := range devices {
				f, err := os.Open(device)
				if err == nil {
					defer f.Close()

					_, err = f.Read(data)
					if err == nil && string(data) == BOOT2DOCKER_MAGIC {
						boot2DockerEnvironment = true
						cfg.Rancher.State.Dev = "LABEL=B2D_STATE"
						cfg.Rancher.State.Autoformat = []string{device}
						break
					}
				}
			}

			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			var err error
			cfg, shouldSwitchRoot, err = tryMountAndBootstrap(cfg)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if err := os.MkdirAll(config.CloudConfigDir, os.ModeDir|0755); err != nil {
				log.Error(err)
			}

			cfg.Rancher.CloudInit.Datasources = config.LoadConfigWithPrefix(STATE).Rancher.CloudInit.Datasources
			if err := config.Set("rancher.cloud_init.datasources", cfg.Rancher.CloudInit.Datasources); err != nil {
				log.Error(err)
			}

			network := false
			for _, datasource := range cfg.Rancher.CloudInit.Datasources {
				if cloudinitsave.RequiresNetwork(datasource) {
					network = true
					break
				}
			}

			if network {
				if err := runCloudInitServices(cfg); err != nil {
					log.Error(err)
				}
			} else {
				if err := cloudinitsave.MountConfigDrive(); err != nil {
					log.Error(err)
				}
				if err := cloudinitsave.SaveCloudConfig(false); err != nil {
					log.Error(err)
				}
				if err := cloudinitsave.UnmountConfigDrive(); err != nil {
					log.Error(err)
				}
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			var err error
			cloudConfigBootFile, err = ioutil.ReadFile(config.CloudConfigBootFile)
			if err != nil {
				log.Error(err)
			}
			metadataFile, err = ioutil.ReadFile(config.MetaDataFile)
			if err != nil {
				log.Error(err)
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if !shouldSwitchRoot {
				return cfg, nil
			}
			log.Debugf("Switching to new root at %s %s", STATE, cfg.Rancher.State.Directory)
			if err := switchRoot(STATE, cfg.Rancher.State.Directory, cfg.Rancher.RmUsr); err != nil {
				return cfg, err
			}
			return cfg, nil
		},
		mountOem,
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if err := os.MkdirAll(config.CloudConfigDir, os.ModeDir|0755); err != nil {
				log.Error(err)
			}
			if err := util.WriteFileAtomic(config.CloudConfigBootFile, cloudConfigBootFile, 400); err != nil {
				log.Error(err)
			}
			if err := util.WriteFileAtomic(config.MetaDataFile, metadataFile, 400); err != nil {
				log.Error(err)
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if boot2DockerEnvironment {
				if err := config.Set("rancher.state.dev", cfg.Rancher.State.Dev); err != nil {
					log.Errorf("Failed to update rancher.state.dev: %v", err)
				}
				if err := config.Set("rancher.state.autoformat", cfg.Rancher.State.Autoformat); err != nil {
					log.Errorf("Failed to update rancher.state.autoformat: %v", err)
				}
			}

			return config.LoadConfig(), nil
		},
		loadModules,
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dfs.PrepareFs(&mountConfig)
		},
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			network.SetProxyEnvironmentVariables(c)
			return c, nil
		},
		initializeSelinux,
		setupSharedRoot,
		sysInit,
	}

	cfg, err := config.ChainCfgFuncs(nil, initFuncs...)
	if err != nil {
		return err
	}

	launchConfig, args := getLaunchConfig(cfg, &cfg.Rancher.SystemDocker)
	launchConfig.Fork = !cfg.Rancher.SystemDocker.Exec

	log.Info("Launching System Docker")
	_, err = dfs.LaunchDocker(launchConfig, config.SYSTEM_DOCKER_BIN, args...)
	if err != nil {
		return err
	}

	return pidOne()
}
