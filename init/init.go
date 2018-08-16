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

	if util.GetHypervisor() == "hyperv" {
		cfg.Rancher.Modules = append(cfg.Rancher.Modules, "hv_utils")
	}

	for _, module := range cfg.Rancher.Modules {
		if mounted[module] {
			continue
		}

		log.Debugf("Loading module %s", module)
		// split module and module parameters
		cmdParam := strings.SplitN(module, " ", -1)
		cmd := exec.Command("modprobe", cmdParam...)
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
	log.InitLogger()
	// TODO: this breaks and does nothing if the cfg is invalid (or is it due to threading?)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Starting Recovery console: %v\n", r)
			recovery(nil)
		}
	}()

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
	err := mountState(cfg)
	if err == nil {
		return nil
	}
	log.Infof("Skipped an error when first mounting: %v", err)

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

func PrintConfig() {
	cfgString, err := config.Export(false, true)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error serializing config")
	} else {
		log.Debugf("Config: %s", cfgString)
	}
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
	hypervisor := ""

	configFiles := make(map[string][]byte)

	initFuncs := []config.CfgFuncData{
		config.CfgFuncData{"preparefs", func(c *config.CloudConfig) (*config.CloudConfig, error) {
			return c, dfs.PrepareFs(&mountConfig)
		}},
		config.CfgFuncData{"save init cmdline", func(c *config.CloudConfig) (*config.CloudConfig, error) {
			// the Kernel Patch added for RancherOS passes `--` (only) elided kernel boot params to the init process
			cmdLineArgs := strings.Join(os.Args, " ")
			config.SaveInitCmdline(cmdLineArgs)

			cfg := config.LoadConfig()
			log.Debugf("Cmdline debug = %t", cfg.Rancher.Debug)
			if cfg.Rancher.Debug {
				log.SetLevel(log.DebugLevel)
			} else {
				log.SetLevel(log.InfoLevel)
			}

			return cfg, nil
		}},
		config.CfgFuncData{"mount OEM", mountOem},
		config.CfgFuncData{"debug save cfg", func(_ *config.CloudConfig) (*config.CloudConfig, error) {
			PrintConfig()

			cfg := config.LoadConfig()
			return cfg, nil
		}},
		config.CfgFuncData{"load modules", loadModules},
		config.CfgFuncData{"recovery console", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if cfg.Rancher.Recovery {
				recovery(nil)
			}
			return cfg, nil
		}},
		config.CfgFuncData{"cloud-init", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			cfg.Rancher.CloudInit.Datasources = config.LoadConfigWithPrefix(state).Rancher.CloudInit.Datasources
			hypervisor = util.GetHypervisor()
			if hypervisor == "" {
				log.Infof("ros init: No Detected Hypervisor")
			} else {
				log.Infof("ros init: Detected Hypervisor: %s", hypervisor)
			}
			if hypervisor == "vmware" {
				// add vmware to the end - we don't want to over-ride an choices the user has made
				cfg.Rancher.CloudInit.Datasources = append(cfg.Rancher.CloudInit.Datasources, hypervisor)
			}

			if err := config.Set("rancher.cloud_init.datasources", cfg.Rancher.CloudInit.Datasources); err != nil {
				log.Error(err)
			}

			log.Infof("init, runCloudInitServices(%v)", cfg.Rancher.CloudInit.Datasources)
			if err := runCloudInitServices(cfg); err != nil {
				log.Error(err)
			}

			// It'd be nice to push to rsyslog before this, but we don't have network
			log.AddRSyslogHook()

			return config.LoadConfig(), nil
		}},
		config.CfgFuncData{"b2d env", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if _, err := os.Stat("/var/lib/boot2docker"); os.IsNotExist(err) {
				err := os.Mkdir("/var/lib/boot2docker", 0755)
				if err != nil {
					log.Errorf("Failed to create boot2docker directory: %v", err)
				}
			}

			if dev := util.ResolveDevice("LABEL=B2D_STATE"); dev != "" {
				boot2DockerEnvironment = true
				cfg.Rancher.State.Dev = "LABEL=B2D_STATE"
				log.Infof("boot2DockerEnvironment %s: %s", cfg.Rancher.State.Dev, dev)
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
						log.Infof("boot2DockerEnvironment %s: Autoformat %s", cfg.Rancher.State.Dev, cfg.Rancher.State.Autoformat[0])

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
		}},
		config.CfgFuncData{"mount and bootstrap", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			var err error
			cfg, shouldSwitchRoot, err = tryMountAndBootstrap(cfg)

			if err != nil {
				return nil, err
			}
			return cfg, nil
		}},
		config.CfgFuncData{"read cfg and log files", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			filesToCopy := []string{
				config.CloudConfigInitFile,
				config.CloudConfigScriptFile,
				config.CloudConfigBootFile,
				config.CloudConfigNetworkFile,
				config.MetaDataFile,
				config.EtcResolvConfFile,
			}
			// And all the files in /var/log/boot/
			// TODO: I wonder if we can put this code into the log module, and have things write to the buffer until we FsReady()
			bootLog := "/var/log/"
			if files, err := ioutil.ReadDir(bootLog); err == nil {
				for _, file := range files {
					if !file.IsDir() {
						filePath := filepath.Join(bootLog, file.Name())
						filesToCopy = append(filesToCopy, filePath)
						log.Debugf("Swizzle: Found %s to save", filePath)
					}
				}
			}
			bootLog = "/var/log/boot/"
			if files, err := ioutil.ReadDir(bootLog); err == nil {
				for _, file := range files {
					filePath := filepath.Join(bootLog, file.Name())
					filesToCopy = append(filesToCopy, filePath)
					log.Debugf("Swizzle: Found %s to save", filePath)
				}
			}
			for _, name := range filesToCopy {
				if _, err := os.Lstat(name); !os.IsNotExist(err) {
					content, err := ioutil.ReadFile(name)
					if err != nil {
						log.Errorf("read cfg file (%s) %s", name, err)
						continue
					}
					log.Debugf("Swizzle: Saved %s to memory", name)
					configFiles[name] = content
				}
			}
			return cfg, nil
		}},
		config.CfgFuncData{"switchroot", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			if !shouldSwitchRoot {
				return cfg, nil
			}
			log.Debugf("Switching to new root at %s %s", state, cfg.Rancher.State.Directory)
			if err := switchRoot(state, cfg.Rancher.State.Directory, cfg.Rancher.RmUsr); err != nil {
				return cfg, err
			}
			return cfg, nil
		}},
		config.CfgFuncData{"mount OEM2", mountOem},
		config.CfgFuncData{"write cfg and log files", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			for name, content := range configFiles {
				dirMode := os.ModeDir | 0755
				fileMode := os.FileMode(0444)
				if strings.HasPrefix(name, "/var/lib/rancher/conf/") {
					// only make the conf files harder to get to
					dirMode = os.ModeDir | 0700
					if name == config.CloudConfigScriptFile {
						fileMode = os.FileMode(0755)
					} else {
						fileMode = os.FileMode(0400)
					}
				}
				if err := os.MkdirAll(filepath.Dir(name), dirMode); err != nil {
					log.Error(err)
				}
				if err := util.WriteFileAtomic(name, content, fileMode); err != nil {
					log.Error(err)
				}
				log.Infof("Swizzle: Wrote file to %s", name)
			}
			if err := os.MkdirAll(config.VarRancherDir, os.ModeDir|0755); err != nil {
				log.Error(err)
			}
			if err := os.Chmod(config.VarRancherDir, os.ModeDir|0755); err != nil {
				log.Error(err)
			}
			log.FsReady()
			log.Debugf("WARNING: switchroot and mount OEM2 phases not written to log file")

			return cfg, nil
		}},
		config.CfgFuncData{"b2d Env", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {

			log.Debugf("memory Resolve.conf == [%s]", configFiles["/etc/resolv.conf"])

			// this code make sure the open-vm-tools service can be started correct when there is no network
			if hypervisor == "vmware" {
				// make sure the cache directory exist
				if err := os.MkdirAll("/var/lib/rancher/cache/", os.ModeDir|0755); err != nil {
					log.Errorf("Create service cache diretory error: %v", err)
				}

				// move os-services cache file
				if _, err := os.Stat("/usr/share/ros/services-cache"); err == nil {
					files, err := ioutil.ReadDir("/usr/share/ros/services-cache/")
					if err != nil {
						log.Errorf("Read file error: %v", err)
					}
					for _, f := range files {
						err := os.Rename("/usr/share/ros/services-cache/"+f.Name(), "/var/lib/rancher/cache/"+f.Name())
						if err != nil {
							log.Errorf("Rename file error: %v", err)
						}
					}
					if err := os.Remove("/usr/share/ros/services-cache"); err != nil {
						log.Errorf("Remove file error: %v", err)
					}
				}

			}

			if boot2DockerEnvironment {
				if err := config.Set("rancher.state.dev", cfg.Rancher.State.Dev); err != nil {
					log.Errorf("Failed to update rancher.state.dev: %v", err)
				}
				if err := config.Set("rancher.state.autoformat", cfg.Rancher.State.Autoformat); err != nil {
					log.Errorf("Failed to update rancher.state.autoformat: %v", err)
				}
			}

			return config.LoadConfig(), nil
		}},
		config.CfgFuncData{"hypervisor tools", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			enableHypervisorService(cfg, hypervisor)
			return config.LoadConfig(), nil
		}},
		config.CfgFuncData{"preparefs2", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			return cfg, dfs.PrepareFs(&mountConfig)
		}},
		config.CfgFuncData{"load modules2", loadModules},
		config.CfgFuncData{"set proxy env", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			network.SetProxyEnvironmentVariables()

			return cfg, nil
		}},
		config.CfgFuncData{"init SELinux", initializeSelinux},
		config.CfgFuncData{"setupSharedRoot", setupSharedRoot},
		config.CfgFuncData{"sysinit", sysInit},
	}

	cfg, err := config.ChainCfgFuncs(nil, initFuncs)
	if err != nil {
		recovery(err)
	}

	launchConfig, args := getLaunchConfig(cfg, &cfg.Rancher.SystemDocker)
	launchConfig.Fork = !cfg.Rancher.SystemDocker.Exec
	//launchConfig.NoLog = true

	log.Info("Launching System Docker")
	_, err = dfs.LaunchDocker(launchConfig, config.SystemDockerBin, args...)
	if err != nil {
		log.Errorf("Error Launching System Docker: %s", err)
		recovery(err)
		return err
	}
	// Code never gets here - rancher.system_docker.exec=true

	return pidOne()
}

func enableHypervisorService(cfg *config.CloudConfig, hypervisorName string) {
	if hypervisorName == "" {
		return
	}

	// only enable open-vm-tools for vmware
	// these services(xenhvm-vm-tools, kvm-vm-tools, hyperv-vm-tools and bhyve-vm-tools) don't exist yet
	serviceName := ""
	switch hypervisorName {
	case "vmware":
		serviceName = "open-vm-tools"
	case "hyperv":
		serviceName = "hyperv-vm-tools"
	default:
		log.Infof("no hypervisor matched")
	}

	if serviceName != "" {
		if !cfg.Rancher.HypervisorService {
			log.Infof("Skipping %s as `rancher.hypervisor_service` is set to false", serviceName)
			return
		}

		// Check removed - there's an x509 cert failure on first boot of an installed system
		// check quickly to see if there is a yml file available
		//	if service.ValidService(serviceName, cfg) {
		log.Infof("Setting rancher.services_include. %s=true", serviceName)
		if err := config.Set("rancher.services_include."+serviceName, "true"); err != nil {
			log.Error(err)
		}
	}
}
