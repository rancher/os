// +build linux

package init

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/mount"
	"github.com/rancher/os/cmd/control"
	networkCmd "github.com/rancher/os/cmd/network"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

const (
	state            string = "/state"
	boot2DockerMagic string = "boot2docker, please format-me"

	tmpfsMagic int64 = 0x01021994
	ramfsMagic int64 = 0x858458f6
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
		cmd := exec.Command("modprobe", module)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Could not load module %s, err %v", module, err)
		}
	}

	return cfg, nil
}

func sysInit(c *config.CloudConfig) (*config.CloudConfig, error) {
	args := append([]string{config.SysInitBin}, os.Args[1:]...)

	cmd := &exec.Cmd{
		Path: config.RosBin,
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
	// TODO: log.InitLogger()
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
	return mountConfigured("state", cfg.Rancher.State.Dev, cfg.Rancher.State.FsType, state)
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

	launchConfig.DNSConfig.Nameservers = cfg.Rancher.Defaults.Network.DNS.Nameservers
	launchConfig.DNSConfig.Search = cfg.Rancher.Defaults.Network.DNS.Search
	launchConfig.Environment = dockerCfg.Environment

	if !cfg.Rancher.Debug {
		launchConfig.LogFile = config.SystemDockerLog
	}

	return &launchConfig, args
}

func isInitrd() bool {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	return int64(stat.Type) == tmpfsMagic || int64(stat.Type) == ramfsMagic
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

	configFiles := make(map[string][]byte)

	initFuncs := []config.CfgFunc{
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dfs.PrepareFs(&mountConfig)
		},
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			// will this be passed to cloud-init-save?
			cmdLineArgs := strings.Join(os.Args, " ")
			config.SaveInitCmdline(cmdLineArgs)

			return c, nil
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
			data := make([]byte, len(boot2DockerMagic))

			for _, device := range devices {
				f, err := os.Open(device)
				if err == nil {
					defer f.Close()

					_, err = f.Read(data)
					if err == nil && string(data) == boot2DockerMagic {
						boot2DockerEnvironment = true
						cfg.Rancher.State.Dev = "LABEL=B2D_STATE"
						cfg.Rancher.State.Autoformat = []string{device}
						break
					}
				}
			}

			// save here so the bootstrap service can see it (when booting from iso, its very early)
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
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			var err error
			cfg, shouldSwitchRoot, err = tryMountAndBootstrap(cfg)
			if err != nil {
				return nil, err
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {

			cfg.Rancher.CloudInit.Datasources = config.LoadConfigWithPrefix(state).Rancher.CloudInit.Datasources
			if err := config.Set("rancher.cloud_init.datasources", cfg.Rancher.CloudInit.Datasources); err != nil {
				log.Error(err)
			}

			if err := control.UdevSettle(); err != nil {
				log.Errorf("Failed to run udev settle: %v", err)
			}

			//cfg := rancherConfig.LoadConfig()
			log.Debugf("init: SaveCloudConfig(pre ApplyNetworkConfig): %#v", cfg.Rancher.Network)
			networkCmd.ApplyNetworkConfig(cfg)

			log.Debug("init: runCloudInitServices()")
			if err := runCloudInitServices(cfg); err != nil {
				log.Error(err)
			}

			// Apply any newly detected network config.
			cfg = config.LoadConfig()
			log.Debugf("init: SaveCloudConfig(post ApplyNetworkConfig): %#v", cfg.Rancher.Network)
			networkCmd.ApplyNetworkConfig(cfg)

			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			filesToCopy := []string{
				config.CloudConfigInitFile,
				config.CloudConfigBootFile,
				config.CloudConfigNetworkFile,
				config.MetaDataFile,
			}
			for _, name := range filesToCopy {
				content, err := ioutil.ReadFile(name)
				if err != nil {
					log.Error(err)
					continue
				}
				configFiles[name] = content
			}
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if !shouldSwitchRoot {
				return cfg, nil
			}
			log.Debugf("Switching to new root at %s %s", state, cfg.Rancher.State.Directory)
			if err := switchRoot(state, cfg.Rancher.State.Directory, cfg.Rancher.RmUsr); err != nil {
				return cfg, err
			}
			return cfg, nil
		},
		mountOem,
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			for name, content := range configFiles {
				if err := os.MkdirAll(filepath.Dir(name), os.ModeDir|0700); err != nil {
					log.Error(err)
				}
				if err := util.WriteFileAtomic(name, content, 400); err != nil {
					log.Error(err)
				}
			}
			if err := os.MkdirAll(config.VarRancherDir, os.ModeDir|0755); err != nil {
				log.Error(err)
			}
			if err := os.Chmod(config.VarRancherDir, os.ModeDir|0755); err != nil {
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
		func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dfs.PrepareFs(&mountConfig)
		},
		loadModules,
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
	_, err = dfs.LaunchDocker(launchConfig, config.SystemDockerBin, args...)
	if err != nil {
		return err
	}

	return pidOne()
}
