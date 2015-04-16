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
		[]string{"devtmpfs", "/dev", "devtmpfs", ""},
		[]string{"none", "/dev/pts", "devpts", ""},
		[]string{"none", "/etc/docker", "tmpfs", ""},
		[]string{"none", "/proc", "proc", ""},
		[]string{"none", "/run", "tmpfs", ""},
		[]string{"none", "/sys", "sysfs", ""},
		[]string{"none", "/sys/fs/cgroup", "tmpfs", ""},
	}
	postMounts [][]string = [][]string{
		[]string{"none", "/var/run", "tmpfs", ""},
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

func createSymlinks(cfg *config.Config, symlinks map[string]string) error {
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

func remountRo(cfg *config.Config) error {
	log.Info("Remouting root read only")
	return util.Remount("/", "ro")
}

func mountCgroups(cfg *config.Config) error {
	for _, cgroup := range cgroups {
		err := createDirs("/sys/fs/cgroup/" + cgroup)
		if err != nil {
			return err
		}

		err = createMounts([][]string{
			[]string{"none", "sys/fs/cgroup/" + cgroup, "cgroup", cgroup},
		}...)
		if err != nil {
			return err
		}
	}

	log.Debug("Done mouting cgroupfs")

	return nil
}

func extractModules(cfg *config.Config) error {
	if _, err := os.Stat(config.MODULES_ARCHIVE); os.IsNotExist(err) {
		log.Debug("Modules do not exist")
		return nil
	}

	log.Debug("Extracting modules")
	return util.ExtractTar(config.MODULES_ARCHIVE, "/")
}

func setResolvConf(cfg *config.Config) error {
	log.Debug("Creating /etc/resolv.conf")
	//f, err := os.OpenFile("/etc/resolv.conf", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	f, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}

	defer f.Close()

	for _, dns := range cfg.Network.Dns.Nameservers {
		content := fmt.Sprintf("nameserver %s\n", dns)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	search := strings.Join(cfg.Network.Dns.Search, " ")
	if search != "" {
		content := fmt.Sprintf("search %s\n", search)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	if cfg.Network.Dns.Domain != "" {
		content := fmt.Sprintf("domain %s\n", cfg.Network.Dns.Domain)
		if _, err = f.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

func loadModules(cfg *config.Config) error {
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

	for _, module := range cfg.Modules {
		log.Debugf("Loading module %s", module)
		err = exec.Command("/sbin/modprobe", module).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func sysInit(cfg *config.Config) error {
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

func execDocker(cfg *config.Config) error {
	log.Info("Launching System Docker")
	if !cfg.Debug {
		output, err := os.Create("/var/log/system-docker.log")
		if err != nil {
			return err
		}

		syscall.Dup2(int(output.Fd()), int(os.Stdout.Fd()))
		syscall.Dup2(int(output.Fd()), int(os.Stderr.Fd()))
	}

	os.Stdin.Close()
	return syscall.Exec(SYSTEM_DOCKER, cfg.SystemDocker.Args, os.Environ())
}

func MainInit() {
	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func mountState(cfg *config.Config) error {
	var err error

	if cfg.State.Dev != "" {
		dev := util.ResolveDevice(cfg.State.Dev)
		log.Infof("Mounting state device %s to %s", dev, STATE)

		fsType := cfg.State.FsType
		if fsType == "auto" {
			fsType, err = util.GetFsType(dev)
		}

		if err == nil {
			log.Debugf("FsType has been set to %s", fsType)
			err = util.Mount(dev, STATE, fsType, "")
		}
	}

	if err != nil && cfg.State.Required {
		return err
	}

	if err != nil || cfg.State.Dev == "" {
		log.Debugf("State will not be persisted")
		err = util.Mount("none", STATE, "tmpfs", "")
	}

	return err
}

func createGroups(cfg *config.Config) error {
	return ioutil.WriteFile("/etc/group", []byte("root:x:0:\n"), 0644)
}

func touchSocket(cfg *config.Config) error {
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

func setupSystemBridge(cfg *config.Config) error {
	bridge, cidr := cfg.SystemDocker.BridgeConfig()
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
	var cfg config.Config

	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")
	os.Setenv("DOCKER_RAMDISK", "true")

	initFuncs := []config.InitFunc{
		func(cfg *config.Config) error {
			return createDirs(dirs...)
		},
		func(cfg *config.Config) error {
			log.Info("Setting up mounts")
			return createMounts(mounts...)
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
		mountCgroups,
		func(cfg *config.Config) error {
			return createSymlinks(cfg, symlinks)
		},
		createGroups,
		extractModules,
		loadModules,
		setResolvConf,
		setupSystemBridge,
		bootstrap,
		mountState,
		func(cfg *config.Config) error {
			return cfg.Reload()
		},
		setResolvConf,
		func(cfg *config.Config) error {
			return createDirs(postDirs...)
		},
		func(cfg *config.Config) error {
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
