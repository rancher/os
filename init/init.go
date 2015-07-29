// +build linux

package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/cmd/network"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	STATE         string = "/var"
	SYSTEM_DOCKER string = "/usr/bin/system-docker"
	DOCKER        string = "/usr/bin/docker"
	SYSINIT       string = "/sbin/rancher-sysinit"
)

var (
	dirs []string = []string{
		"/etc/ssl/certs",
		"/sbin",
		"/usr/bin",
		"/usr/sbin",
	}
	postDirs []string = []string{
		"/var/log",
		"/var/lib/rancher/state/home",
		"/var/lib/rancher/state/opt",
	}
	mounts [][]string = [][]string{
		{"devtmpfs", "/dev", "devtmpfs", ""},
		{"none", "/dev/pts", "devpts", ""},
		{"none", "/etc/docker", "tmpfs", ""},
		{"none", "/proc", "proc", ""},
		{"none", "/run", "tmpfs", ""},
		{"none", "/sys", "sysfs", ""},
		{"none", "/sys/fs/cgroup", "tmpfs", ""},
	}
	postMounts [][]string = [][]string{
		{"none", "/var/run", "tmpfs", ""},
	}
	cgroups []string = []string{
		"blkio",
		"cpu",
		"cpuacct",
		"cpuset",
		"devices",
		"freezer",
		"memory",
		"net_cls",
		"perf_event",
	}
	// Notice this map is the reverse order of a "ln -s x y" command
	// so map[y] = x
	symlinks map[string]string = map[string]string{
		"/etc/ssl/certs/ca-certificates.crt": "/ca.crt",
		"/sbin/modprobe":                     "/busybox",
		"/usr/sbin/iptables":                 "/xtables-multi",
		DOCKER:                               "/docker",
		SYSTEM_DOCKER:                        "/docker",
		SYSINIT:                              "/init",
		"/home":                              "/var/lib/rancher/state/home",
		"/opt":                               "/var/lib/rancher/state/opt",
	}
)

func createSymlinks(cfg *config.CloudConfig, symlinks map[string]string) error {
	log.Debug("Creating symlinking")
	for dest, src := range symlinks {
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			log.Debugf("Symlinking %s => %s", src, dest)
			if err = os.Symlink(src, dest); err != nil {
				return err
			}
		}
	}

	return nil
}

func createDirs(dirs ...string) error {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Debugf("Creating %s", dir)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createMounts(mounts ...[]string) error {
	for _, mount := range mounts {
		log.Debugf("Mounting %s %s %s %s", mount[0], mount[1], mount[2], mount[3])
		err := util.Mount(mount[0], mount[1], mount[2], mount[3])
		if err != nil {
			return err
		}
	}

	return nil
}

func remountRo(cfg *config.CloudConfig) error {
	log.Info("Remouting root read only")
	return util.Remount("/", "ro")
}

func mountCgroups(cfg *config.CloudConfig) error {
	for _, cgroup := range cgroups {
		err := createDirs("/sys/fs/cgroup/" + cgroup)
		if err != nil {
			return err
		}

		err = createMounts([][]string{
			{"none", "sys/fs/cgroup/" + cgroup, "cgroup", cgroup},
		}...)
		if err != nil {
			return err
		}
	}

	log.Debug("Done mouting cgroupfs")

	return nil
}

func extractModules(cfg *config.CloudConfig) error {
	if _, err := os.Stat(config.MODULES_ARCHIVE); os.IsNotExist(err) {
		log.Debug("Modules do not exist")
		return nil
	}

	log.Debug("Extracting modules")
	return util.ExtractTar(config.MODULES_ARCHIVE, "/")
}

func setResolvConf(cfg *config.CloudConfig) error {
	log.Debug("Creating /etc/resolv.conf")
	//f, err := os.OpenFile("/etc/resolv.conf", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	f, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}

	defer f.Close()

	for _, dns := range cfg.Rancher.Network.Dns.Nameservers {
		content := fmt.Sprintf("nameserver %s\n", dns)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	search := strings.Join(cfg.Rancher.Network.Dns.Search, " ")
	if search != "" {
		content := fmt.Sprintf("search %s\n", search)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	if cfg.Rancher.Network.Dns.Domain != "" {
		content := fmt.Sprintf("domain %s\n", cfg.Rancher.Network.Dns.Domain)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func loadModules(cfg *config.CloudConfig) error {
	filesystems, err := ioutil.ReadFile("/proc/filesystems")
	if err != nil {
		return err
	}

	if !strings.Contains(string(filesystems), "nodev\toverlay\n") {
		log.Debug("Loading overlay module")
		err = exec.Command("/sbin/modprobe", "overlay").Run()
		if err != nil {
			return err
		}
	}

	for _, module := range cfg.Rancher.Modules {
		log.Debugf("Loading module %s", module)
		err = exec.Command("/sbin/modprobe", module).Run()
		if err != nil {
			log.Errorf("Could not load module %s, err %v", module, err)
		}
	}

	return nil
}

func sysInit(cfg *config.CloudConfig) error {
	args := append([]string{SYSINIT}, os.Args[1:]...)

	var cmd *exec.Cmd
	if util.IsRunningInTty() {
		cmd = exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return os.Stdin.Close()
}

func execDocker(cfg *config.CloudConfig) error {
	log.Info("Launching System Docker")
	if !cfg.Rancher.Debug {
		output, err := os.Create("/var/log/system-docker.log")
		if err != nil {
			return err
		}

		syscall.Dup2(int(output.Fd()), int(os.Stdout.Fd()))
		syscall.Dup2(int(output.Fd()), int(os.Stderr.Fd()))
	}

	os.Stdin.Close()
	return syscall.Exec(SYSTEM_DOCKER, cfg.Rancher.SystemDocker.Args, os.Environ())
}

func MainInit() {
	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func mountStateTmpfs(cfg *config.CloudConfig) error {
	log.Debugf("State will not be persisted")
	return util.Mount("none", STATE, "tmpfs", "")
}

func mountState(cfg *config.CloudConfig) error {
	var err error

	if cfg.Rancher.State.Dev != "" {
		dev := util.ResolveDevice(cfg.Rancher.State.Dev)
		if dev == "" {
			msg := fmt.Sprintf("Could not resolve device %q", cfg.Rancher.State.Dev)
			log.Infof(msg)
			return fmt.Errorf(msg)
		}
		log.Infof("Mounting state device %s to %s", dev, STATE)

		fsType := cfg.Rancher.State.FsType
		if fsType == "auto" {
			fsType, err = util.GetFsType(dev)
		}

		if err == nil {
			log.Debugf("FsType has been set to %s", fsType)
			err = util.Mount(dev, STATE, fsType, "")
		}
	} else {
		return mountStateTmpfs(cfg)
	}

	return err
}

func tryMountAndBootstrap(cfg *config.CloudConfig) error {
	if err := mountState(cfg); err != nil {
		if err := bootstrap(cfg); err != nil {
			if cfg.Rancher.State.Required {
				return err
			}
			return mountStateTmpfs(cfg)
		}
		if err := mountState(cfg); err != nil {
			if cfg.Rancher.State.Required {
				return err
			}
			return mountStateTmpfs(cfg)
		}
	}
	return nil
}

func createGroups(cfg *config.CloudConfig) error {
	return ioutil.WriteFile("/etc/group", []byte("root:x:0:\n"), 0644)
}

func touchSocket(cfg *config.CloudConfig) error {
	for _, path := range []string{"/var/run/docker.sock", "/var/run/system-docker.sock"} {
		if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		err := ioutil.WriteFile(path, []byte{}, 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupSystemBridge(cfg *config.CloudConfig) error {
	bridge, cidr := cfg.Rancher.SystemDocker.BridgeConfig()
	if bridge == "" {
		return nil
	}

	return network.ApplyNetworkConfigs(&config.NetworkConfig{
		Interfaces: map[string]config.InterfaceConfig{
			bridge: {
				Bridge:  true,
				Address: cidr,
			},
		},
	})
}

func RunInit() error {
	var cfg config.CloudConfig

	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")
	os.Setenv("DOCKER_RAMDISK", "true")

	initFuncs := []config.InitFunc{
		func(cfg *config.CloudConfig) error {
			return createDirs(dirs...)
		},
		func(cfg *config.CloudConfig) error {
			log.Info("Setting up mounts")
			return createMounts(mounts...)
		},
		func(cfg *config.CloudConfig) error {
			newCfg, err := config.LoadConfig()
			if err == nil {
				newCfg, err = config.LoadConfig()
			}
			if err == nil {
				*cfg = *newCfg
			}

			if cfg.Rancher.Debug {
				cfgString, _ := config.Dump(false, true)
				log.Debugf("os-config dump: \n%s", cfgString)
			}

			return err
		},
		mountCgroups,
		func(cfg *config.CloudConfig) error {
			return createSymlinks(cfg, symlinks)
		},
		createGroups,
		extractModules,
		loadModules,
		setResolvConf,
		setupSystemBridge,
		tryMountAndBootstrap,
		func(cfg *config.CloudConfig) error {
			return cfg.Reload()
		},
		loadModules,
		setResolvConf,
		func(cfg *config.CloudConfig) error {
			return createDirs(postDirs...)
		},
		func(cfg *config.CloudConfig) error {
			return createMounts(postMounts...)
		},
		touchSocket,
		// Disable R/O root write now to support updating modules
		//remountRo,
		sysInit,
	}

	if err := config.RunInitFuncs(&cfg, initFuncs); err != nil {
		return err
	}

	return execDocker(&cfg)
}
