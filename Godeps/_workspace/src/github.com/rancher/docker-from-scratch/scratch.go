package dockerlaunch

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/resolvconf"
	"github.com/rancher/docker-from-scratch/util"
	"github.com/rancher/netconf"
)

const defaultPrefix = "/usr"

var (
	mounts [][]string = [][]string{
		{"devtmpfs", "/dev", "devtmpfs", ""},
		{"none", "/dev/pts", "devpts", ""},
		{"none", "/proc", "proc", ""},
		{"none", "/run", "tmpfs", ""},
		{"none", "/sys", "sysfs", ""},
		{"none", "/sys/fs/cgroup", "tmpfs", ""},
	}
)

type Config struct {
	Fork            bool
	CommandName     string
	DnsConfig       netconf.DnsConfig
	BridgeName      string
	BridgeAddress   string
	BridgeMtu       int
	CgroupHierarchy map[string]string
	LogFile         string
	NoLog           bool
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

func mountCgroups(hierarchyConfig map[string]string) error {
	f, err := os.Open("/proc/cgroups")
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	hierarchies := make(map[string][]string)

	for scanner.Scan() {
		text := scanner.Text()
		log.Debugf("/proc/cgroups: %s", text)
		fields := strings.SplitN(text, "\t", 3)
		cgroup := fields[0]
		if cgroup == "" || cgroup[0] == '#' || len(fields) < 3 {
			continue
		}

		hierarchy := hierarchyConfig[cgroup]
		if hierarchy == "" {
			hierarchy = fields[1]
		}

		if hierarchy == "0" {
			hierarchy = cgroup
		}

		hierarchies[hierarchy] = append(hierarchies[hierarchy], cgroup)
	}

	for _, hierarchy := range hierarchies {
		if err := mountCgroup(strings.Join(hierarchy, ",")); err != nil {
			return err
		}
	}

	if err = scanner.Err(); err != nil {
		return err
	}

	log.Debug("Done mouting cgroupfs")
	return nil
}

func CreateSymlinks(pathSets [][]string) error {
	for _, paths := range pathSets {
		if err := CreateSymlink(paths[0], paths[1]); err != nil {
			return err
		}
	}

	return nil
}

func CreateSymlink(src, dest string) error {
	if _, err := os.Lstat(dest); os.IsNotExist(err) {
		log.Debugf("Symlinking %s => %s", src, dest)
		if err = os.Symlink(src, dest); err != nil {
			return err
		}
	}

	return nil
}

func mountCgroup(cgroup string) error {
	if err := createDirs("/sys/fs/cgroup/" + cgroup); err != nil {
		return err
	}

	if err := createMounts([][]string{{"none", "/sys/fs/cgroup/" + cgroup, "cgroup", cgroup}}...); err != nil {
		return err
	}

	parts := strings.Split(cgroup, ",")
	if len(parts) > 1 {
		for _, part := range parts {
			if err := CreateSymlink("/sys/fs/cgroup/"+cgroup, "/sys/fs/cgroup/"+part); err != nil {
				return err
			}
		}
	}

	return nil
}

func execDocker(config *Config, docker, cmd string, args []string) (*exec.Cmd, error) {
	if len(args) > 0 && args[0] == "docker" {
		args = args[1:]
	}
	log.Debugf("Launching Docker %s %s %v", docker, cmd, args)

	if config.Fork {
		cmd := exec.Command(docker, args...)
		if !config.NoLog {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		err := cmd.Start()
		return cmd, err
	} else {
		err := syscall.Exec(docker, append([]string{cmd}, args...), os.Environ())
		return nil, err
	}
}

func copyDefault(folder, name string) error {
	defaultFile := path.Join(defaultPrefix, folder, name)
	if err := CopyFile(defaultFile, folder, name); err != nil {
		return err
	}

	return nil
}

func defaultFiles(files ...string) error {
	for _, file := range files {
		dir := path.Dir(file)
		name := path.Base(file)
		if err := copyDefault(dir, name); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(src, folder, name string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	dst := path.Join(folder, name)
	if _, err := os.Stat(dst); err == nil {
		return nil
	}

	if err := createDirs(folder); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func tryCreateFile(name, content string) error {
	if _, err := os.Stat(name); err == nil {
		return nil
	}

	if err := createDirs(path.Dir(name)); err != nil {
		return err
	}

	return ioutil.WriteFile(name, []byte(content), 0644)
}

func createPasswd() error {
	return tryCreateFile("/etc/passwd", "root:x:0:0:root:/root:/bin/sh\n")
}

func createGroup() error {
	return tryCreateFile("/etc/group", "root:x:0:\n")
}

func setupNetworking(config *Config) error {
	if config == nil {
		return nil
	}

	if len(config.DnsConfig.Nameservers) != 0 {
		if err := resolvconf.Build("/etc/resolv.conf", config.DnsConfig.Nameservers, config.DnsConfig.Search); err != nil {
			return err
		}
	}

	if config.BridgeName != "" {
		log.Debugf("Creating bridge %s (%s)", config.BridgeName, config.BridgeAddress)
		if err := netconf.ApplyNetworkConfigs(&netconf.NetworkConfig{
			Interfaces: map[string]netconf.InterfaceConfig{
				config.BridgeName: {
					Address: config.BridgeAddress,
					MTU:     config.BridgeMtu,
					Bridge:  true,
				},
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func getValue(index int, args []string) string {
	val := args[index]
	parts := strings.SplitN(val, "=", 2)
	if len(parts) == 1 {
		if len(args) > index+1 {
			return args[index+1]
		} else {
			return ""
		}
	} else {
		return parts[2]
	}
}

func ParseConfig(config *Config, args ...string) []string {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--bip") {
			config.BridgeAddress = getValue(i, args)
		} else if strings.HasPrefix(arg, "--fixed-cidr") {
			config.BridgeAddress = getValue(i, args)
		} else if strings.HasPrefix(arg, "-b") || strings.HasPrefix(arg, "--bridge") {
			config.BridgeName = getValue(i, args)
		} else if strings.HasPrefix(arg, "--mtu") {
			mtu, err := strconv.Atoi(getValue(i, args))
			if err != nil {
				config.BridgeMtu = mtu
			}
		}
	}

	if config.BridgeName != "" && config.BridgeAddress != "" {
		newArgs := []string{}
		skip := false
		for _, arg := range args {
			if skip {
				skip = false
				continue
			}

			if arg == "--bip" {
				skip = true
				continue
			} else if strings.HasPrefix(arg, "--bip=") {
				continue
			}

			newArgs = append(newArgs, arg)
		}

		args = newArgs
	}

	return args
}

func PrepareFs(config *Config) error {
	if err := createMounts(mounts...); err != nil {
		return err
	}

	if err := mountCgroups(config.CgroupHierarchy); err != nil {
		return err
	}

	if err := createLayout(); err != nil {
		return err
	}

	return nil
}

func touchSocket(path string) error {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return ioutil.WriteFile(path, []byte{}, 0700)
}

func touchSockets(args ...string) error {
	touched := false

	for i, arg := range args {
		if strings.HasPrefix(arg, "-H") {
			val := getValue(i, args)
			if strings.HasPrefix(val, "unix://") {
				val = val[len("unix://"):]
				log.Debugf("Creating temp file at %s", val)
				if err := touchSocket(val); err != nil {
					return err
				}
				touched = true
			}
		}
	}

	if !touched {
		return touchSocket("/var/run/docker.sock")
	}

	return nil
}

func createLayout() error {
	if err := createDirs("/tmp", "/root/.ssh", "/var"); err != nil {
		return err
	}

	return CreateSymlinks([][]string{
		{"usr/lib", "/lib"},
		{"usr/sbin", "/sbin"},
		{"../run", "/var/run"},
	})
}

func prepare(config *Config, docker string, args ...string) error {
	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")

	if err := defaultFiles(
		"/etc/ssl/certs/ca-certificates.crt",
		"/etc/passwd",
		"/etc/group",
	); err != nil {
		return err
	}

	if err := createPasswd(); err != nil {
		return err
	}

	if err := createGroup(); err != nil {
		return err
	}

	if err := setupNetworking(config); err != nil {
		return err
	}

	if err := touchSockets(args...); err != nil {
		return err
	}

	if err := setupLogging(config); err != nil {
		return err
	}

	if err := setupBin(config, docker); err != nil {
		return err
	}

	return nil
}

func setupBin(config *Config, docker string) error {
	if _, err := os.Stat(docker); os.IsNotExist(err) {
		dist := docker + ".dist"
		if _, err := os.Stat(dist); err == nil {
			return os.Symlink(dist, docker)
		}
	}

	return nil
}

func setupLogging(config *Config) error {
	if config.LogFile == "" {
		return nil
	}

	if err := createDirs(path.Dir(config.LogFile)); err != nil {
		return err
	}

	output, err := os.Create(config.LogFile)
	if err != nil {
		return err
	}

	syscall.Dup2(int(output.Fd()), int(os.Stdout.Fd()))
	syscall.Dup2(int(output.Fd()), int(os.Stderr.Fd()))

	return nil
}

func runOrExec(config *Config, docker string, args ...string) (*exec.Cmd, error) {
	if err := prepare(config, docker, args...); err != nil {
		return nil, err
	}

	cmd := "docker"
	if config != nil && config.CommandName != "" {
		cmd = config.CommandName
	}

	return execDocker(config, docker, cmd, args)
}

func LaunchDocker(config *Config, docker string, args ...string) (*exec.Cmd, error) {
	if err := PrepareFs(config); err != nil {
		return nil, err
	}

	return runOrExec(config, docker, args...)
}
