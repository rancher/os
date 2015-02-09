package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	STATE  string = "/var/lib/rancher/state"
	DOCKER string = "/var/lib/docker"
)

var (
	dirs []string = []string{
		"/bin",
		"/dev",
		"/dev/pts",
		"/etc/ssl/certs/ca-certificates.crt",
		"/proc",
		"/sbin",
		"/sys",
		"/usr/bin",
		DOCKER,
		STATE,
		"/var/run",
	}
	mounts [][]string = [][]string{
		[]string{"none", "/etc/docker", "tmpfs", ""},
		[]string{"none", "/proc", "proc", ""},
		[]string{"devtmpfs", "/dev", "devtmpfs", ""},
		[]string{"none", "/sys", "sysfs", ""},
		[]string{"none", "/sys/fs/cgroup", "tmpfs", ""},
		[]string{"none", "/dev/pts", "devpts", ""},
		[]string{"none", "/var/run", "tmpfs", ""},
	}
	cgroups []string = []string{
		"perf_event",
		"net_cls",
		"freezer",
		"devices",
		"blkio",
		"memory",
		"cpuacct",
		"cpu",
		"cpuset",
	}
	// Notice this map is the reverse order of a "ln -s x y" command
	// so map[y] = x
	symlinks map[string]string = map[string]string{
		"/etc/ssl/certs/ca-certificates.crt": "/ca.crt",
		"/bin/busybox":                       "/busybox",
		"/sbin/init-sys":                     "/init",
		"/sbin/init-user":                    "/init",
		"/sbin/modprobe":                     "/bin/busybox",
		"/usr/bin/docker":                    "/docker",
		"/usr/bin/openvt":                    "/busybox",
		"/usr/bin/system-docker":             "/init",
	}
)

func createSymlinks(cfg *config.Config) error {
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
	if _, err := os.Stat(cfg.ModulesArchive); os.IsNotExist(err) {
		log.Debug("Modules do not exist")
		return nil
	}

	log.Debug("Extracting modules")
	return util.ExtractTar(cfg.ModulesArchive, "/")
}

func setResolvConf(cfg *config.Config) error {
	log.Debug("Creating /etc/resolv.conf")
	//f, err := os.OpenFile("/etc/resolv.conf", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	f, err := os.Create("/etc/resolv.conf")
	if err != nil {
		return err
	}

	defer f.Close()

	for _, dns := range cfg.Dns {
		content := fmt.Sprintf("nameserver %s\n", dns)
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
	cmd := exec.Command("openvt", "-s", "/sbin/init-sys")
	//cmd := exec.Command("/sbin/init-sys")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return err
	}

	return nil

	//log.Debug("Launching host console")
	//return exec.Command("openvt", "/bin/sh").Run()

	//log.Debug("Launching console")
	//return exec.Command("/bin/openvt", "-s", "/bin/console-container.sh").Start()
}

func execDocker(cfg *config.Config) error {
	log.Info("Launching Docker")
	return syscall.Exec(cfg.DockerBin, cfg.SystemDockerArgs, os.Environ())
}

func MainInit() {
	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func mountState(cfg *config.Config) error {
	var err error
	if len(cfg.StateDev) == 0 {
		log.Debugf("State will not be persisted")
		err = util.Mount("none", STATE, "tmpfs", "")
	} else {
		log.Debugf("Mounting state device %s", cfg.StateDev)
		err = util.Mount(cfg.StateDev, STATE, cfg.StateDevFSType, "")
	}

	if err != nil {
		return err
	}

	for _, i := range []string{"docker", "images"} {
		dir := path.Join(STATE, i)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}

	log.Debugf("Bind mounting %s to %s", path.Join(STATE, "docker"), DOCKER)
	err = util.Mount(path.Join(STATE, "docker"), DOCKER, "", "bind")
	if err != nil && cfg.StateRequired {
		return err
	}

	return nil
}

func pause(cfg *config.Config) error {
	time.Sleep(5 + time.Minute)
	return nil
}

func RunInit() error {
	var cfg config.Config

	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	os.Setenv("DOCKER_RAMDISK", "true")

	initFuncs := []config.InitFunc{
		func(cfg *config.Config) error {
			return createDirs(dirs...)
		},
		func(cfg *config.Config) error {
			return createMounts(mounts...)
		},
		func(cfg *config.Config) error {
			newCfg, err := config.LoadConfig()
			if err == nil {
				*cfg = *newCfg
			}
			return err
		},
		createSymlinks,
		setResolvConf,
		extractModules,
		remountRo,
		mountCgroups,
		loadModules,
		mountState,
		sysInit,
	}

	if err := config.RunInitFuncs(&cfg, initFuncs); err != nil {
		return err
	}

	return execDocker(&cfg)
}
