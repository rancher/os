package dfs

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"

	"github.com/burmilla/os/config/cmdline"
	"github.com/burmilla/os/pkg/init/one"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/netconf"
	"github.com/burmilla/os/pkg/util"

	"github.com/docker/libnetwork/resolvconf"
)

const (
	defaultPrefix = "/usr"
	iptables      = "/sbin/iptables"
	modprobe      = "/sbin/modprobe"
	distSuffix    = ".dist"
	cgroupRoot    = "/sys/fs/cgroup"
	cgroupV2Fs    = "cgroup2"
)

var (
	mounts = [][]string{
		{"devtmpfs", "/dev", "devtmpfs", ""},
		{"none", "/dev/pts", "devpts", ""},
		{"shm", "/dev/shm", "tmpfs", "rw,nosuid,nodev,noexec,relatime,size=65536k"},
		{"mqueue", "/dev/mqueue", "mqueue", "rw,nosuid,nodev,noexec,relatime"},
		{"none", "/proc", "proc", ""},
		{"none", "/run", "tmpfs", ""},
		{"none", "/sys", "sysfs", ""},
		{"debugfs", "/sys/kernel/debug", "debugfs", ""},
	}

	// The unified hierarchy is mounted directly on /sys/fs/cgroup, it must not
	// be preceded by the tmpfs which the cgroup v1 layout used as a container
	// for the per controller hierarchies.
	cgroupV2Mount = []string{cgroupV2Fs, cgroupRoot, cgroupV2Fs, "rw,nosuid,nodev,noexec,relatime,nsdelegate"}

	// Controllers which are delegated to the first level of the unified
	// hierarchy when the kernel makes them available.
	cgroupV2Controllers = []string{"cpu", "cpuset", "io", "memory", "pids", "hugetlb", "rdma", "misc"}
)

type Config struct {
	Fork              bool
	PidOne            bool
	CommandName       string
	DNSConfig         netconf.DNSConfig
	BridgeName        string
	BridgeAddress     string
	BridgeMtu         int
	LogFile           string
	NoLog             bool
	NoFiles           uint64
	Environment       []string
	DataRootDirectory string
	DaemonConfig      string
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

// mountCgroupV2 mounts the cgroup v2 unified hierarchy on /sys/fs/cgroup and
// delegates the available controllers to the first level of the tree.
func mountCgroupV2() error {
	if err := createDirs(cgroupRoot); err != nil {
		return err
	}

	fsType, err := util.GetMountFsType(cgroupRoot)
	if err != nil {
		return err
	}

	switch fsType {
	case cgroupV2Fs:
		log.Debugf("%s is already a cgroup v2 mount", cgroupRoot)
	case "":
		if err := createMounts(cgroupV2Mount); err != nil {
			return err
		}
	default:
		// Anything else on /sys/fs/cgroup, a leftover tmpfs or a v1
		// hierarchy, would just hide the unified hierarchy, so it has to
		// go before cgroup2 can be mounted.
		log.Infof("Unmounting %s (%s) to mount the cgroup v2 hierarchy", cgroupRoot, fsType)
		if err := util.Unmount(cgroupRoot); err != nil {
			return err
		}
		if err := createMounts(cgroupV2Mount); err != nil {
			return err
		}
	}

	enableCgroupV2Controllers()

	return nil
}

// enableCgroupV2Controllers makes the controllers of the root cgroup available
// to its children. Without this only the processes living in the root cgroup
// itself could be accounted and limited. Failures are not fatal, the kernel
// may simply not have the controller compiled in or it may be in use already.
func enableCgroupV2Controllers() {
	available, err := ioutil.ReadFile(path.Join(cgroupRoot, "cgroup.controllers"))
	if err != nil {
		log.Errorf("Failed to read cgroup v2 controllers: %v", err)
		return
	}

	enabled := map[string]bool{}
	for _, controller := range strings.Fields(string(available)) {
		enabled[controller] = true
	}

	subtreeControl := path.Join(cgroupRoot, "cgroup.subtree_control")
	for _, controller := range cgroupV2Controllers {
		if !enabled[controller] {
			continue
		}
		if err := ioutil.WriteFile(subtreeControl, []byte("+"+controller), 0644); err != nil {
			log.Warnf("Failed to enable cgroup v2 controller %s: %v", controller, err)
			continue
		}
		log.Debugf("Enabled cgroup v2 controller %s", controller)
	}
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
		log.Debugf("Symlinking %s => %s", dest, src)
		if err = os.Symlink(src, dest); err != nil {
			return err
		}
	}

	return nil
}

func execDocker(config *Config, docker, cmd string, args []string) (*exec.Cmd, error) {
	if len(args) > 0 && args[0] == "docker" {
		args = args[1:]
	}
	log.Debugf("Launching Docker %s %s %v", docker, cmd, args)

	env := os.Environ()
	if len(config.Environment) != 0 {
		env = append(env, config.Environment...)
	}

	if config.Fork {
		cmd := exec.Command(docker, args...)
		if !config.NoLog {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		cmd.Env = env
		err := cmd.Start()
		if err != nil {
			return cmd, err
		}
		if config.PidOne {
			one.PidOne()
		}
		return cmd, err
	}

	return nil, syscall.Exec(expand(docker), append([]string{cmd}, args...), env)
}

func copyDefault(folder, name string) error {
	defaultFile := path.Join(defaultPrefix, folder, name)
	return CopyFile(defaultFile, folder, name)
}

func copyDefaultFolder(folder string) error {
	log.Debugf("Copying folder %s", folder)
	defaultFolder := path.Join(defaultPrefix, folder)
	files, _ := ioutil.ReadDir(defaultFolder)
	for _, file := range files {
		var err error
		if file.IsDir() {
			err = copyDefaultFolder(path.Join(folder, file.Name()))
		} else {
			err = copyDefault(folder, file.Name())
		}
		if err != nil {
			return err
		}
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

func defaultFolders(folders ...string) error {
	for _, folder := range folders {
		if err := copyDefaultFolder(folder); err != nil {
			return err
		}
	}

	return nil
}

func CopyFile(src, folder, name string) error {
	return CopyFileOverwrite(src, folder, name, false)
}

func CopyFileOverwrite(src, folder, name string, overwrite bool) error {
	if _, err := os.Lstat(src); os.IsNotExist(err) {
		log.Debugf("Not copying %s, does not exists", src)
		return nil
	}

	dst := path.Join(folder, name)
	if !overwrite {
		if _, err := os.Lstat(dst); err == nil {
			log.Debugf("Not copying %s => %s already exists", src, dst)
			return nil
		}
	}

	if err := createDirs(folder); err != nil {
		return err
	}

	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		symDst, err := os.Readlink(src)
		if err != nil {
			log.Errorf("Failed to readlink: %v", err)
			return err
		}
		// file is a symlink
		log.Debugf("Symlinking %s => %s", dst, symDst)
		return os.Symlink(symDst, dst)
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

	log.Debugf("Copying %s => %s", src, dst)
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

func setupNetworking(cfg *Config) error {
	if cfg == nil {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	tryCreateFile("/etc/hosts", `127.0.0.1    localhost
::1    localhost ip6-localhost ip6-loopback
fe00::0    ip6-localnet
ff00::0    ip6-mcastprefix
ff02::1    ip6-allnodes
ff02::2    ip6-allrouters

127.0.1.1       `+hostname)

	if len(cfg.DNSConfig.Nameservers) != 0 {
		resolve, err := ioutil.ReadFile("/etc/resolv.conf")
		log.Debugf("Resolve.conf == [%s], %v", resolve, err)

		if err != nil {
			log.Infof("scratch Writing empty resolv.conf (%v) %v", []string{}, []string{})
			if _, err := resolvconf.Build("/etc/resolv.conf", []string{}, []string{}, nil); err != nil {
				return err
			}
		}
	}

	if cfg.BridgeName != "" && cfg.BridgeName != "none" {
		log.Debugf("Creating bridge %s (%s)", cfg.BridgeName, cfg.BridgeAddress)
		if _, err := netconf.ApplyNetworkConfigs(&netconf.NetworkConfig{
			Interfaces: map[string]netconf.InterfaceConfig{
				cfg.BridgeName: {
					Address: cfg.BridgeAddress,
					MTU:     cfg.BridgeMtu,
					Bridge:  "true",
				},
			},
		}, false, false); err != nil {
			log.Errorf("Error creating bridge: %s", err)
			return err
		}
	}

	return nil
}

func GetValue(index int, args []string) string {
	val := args[index]
	parts := strings.SplitN(val, "=", 2)
	if len(parts) == 1 {
		if len(args) > index+1 {
			return args[index+1]
		}
		return ""
	}
	return parts[1]
}

func ParseConfig(config *Config, args ...string) []string {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--bip") {
			config.BridgeAddress = GetValue(i, args)
		} else if strings.HasPrefix(arg, "--fixed-cidr") {
			config.BridgeAddress = GetValue(i, args)
		} else if strings.HasPrefix(arg, "-b") || strings.HasPrefix(arg, "--bridge") {
			config.BridgeName = GetValue(i, args)
		} else if strings.HasPrefix(arg, "--config-file") {
			config.DaemonConfig = GetValue(i, args)
		} else if strings.HasPrefix(arg, "--mtu") {
			mtu, err := strconv.Atoi(GetValue(i, args))
			if err != nil {
				config.BridgeMtu = mtu
			}
		} else if strings.HasPrefix(arg, "--data-root") {
			config.DataRootDirectory = GetValue(i, args)
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

	if err := mountCgroupV2(); err != nil {
		return err
	}

	if err := createLayout(config); err != nil {
		return err
	}

	return firstPrepare()
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
			val := GetValue(i, args)
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

func createDaemonConfig(config *Config) error {
	if config.DaemonConfig == "" {
		return nil
	}

	if _, err := os.Stat(config.DaemonConfig); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(config.DaemonConfig), 0755); err != nil {
			return err
		}

		return ioutil.WriteFile(config.DaemonConfig, []byte("{}"), 0600)
	}

	return nil
}

func cleanupFiles(dataRootDirectory string) {
	zeroFiles := []string{
		"/etc/docker/key.json",
		"/etc/docker/daemon.json",
		"/etc/docker/system-daemon.json",
		path.Join(dataRootDirectory, "image/overlay/repositories.json"),
	}

	for _, file := range zeroFiles {
		if stat, err := os.Stat(file); err == nil {
			if stat.Size() < 2 {
				log.Warnf("Deleting invalid json file: %s", file)
				os.Remove(file)
			}
		}
	}
}

func createLayout(config *Config) error {
	if err := createDirs("/tmp", "/root/.ssh", "/var", "/usr/lib"); err != nil {
		return err
	}

	dataRootDirectory := config.DataRootDirectory

	if config.DataRootDirectory == "" {
		dataRootDirectory = "/var/lib/docker"
	}

	if err := createDirs(dataRootDirectory); err != nil {
		return err
	}

	if err := createDaemonConfig(config); err != nil {
		return err
	}

	cleanupFiles(dataRootDirectory)

	symlinks := [][]string{
		{"usr/lib", "/lib"},
		{"usr/sbin", "/sbin"},
		{"../run", "/var/run"},
	}

	rootCmdline := cmdline.GetCmdline("root")
	rootDevice := rootCmdline.(string)
	if rootDevice != "" {
		if _, err := os.Stat("/dev/root"); os.IsNotExist(err) {
			symlinks = append(symlinks, []string{rootDevice, "/dev/root"})
		}
	}

	return CreateSymlinks(symlinks)
}

func firstPrepare() error {
	os.Setenv("PATH", "/sbin:/usr/sbin:/usr/bin")

	if err := defaultFiles(
		"/etc/ssl/certs/ca-certificates.crt",
		"/etc/passwd",
		"/etc/group",
	); err != nil {
		return err
	}

	if err := defaultFolders(
		"/etc/docker",
	); err != nil {
		return err
	}

	if err := createPasswd(); err != nil {
		return err
	}

	return createGroup()
}

func secondPrepare(config *Config, docker string, args ...string) error {

	if err := setupNetworking(config); err != nil {
		return err
	}

	if err := touchSockets(args...); err != nil {
		return err
	}

	if err := setupLogging(config); err != nil {
		return err
	}

	for _, i := range []string{docker, iptables, modprobe} {
		if err := setupBin(config, i); err != nil {
			return err
		}
	}

	if err := setUlimit(config); err != nil {
		return err
	}

	ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0655)

	return nil
}

func expand(bin string) string {
	expanded, err := exec.LookPath(bin)
	if err == nil {
		return expanded
	}
	return bin
}

func setupBin(config *Config, bin string) error {
	expanded, err := exec.LookPath(bin)
	if err == nil {
		return nil
	}

	expanded, err = exec.LookPath(bin + distSuffix)
	if err != nil {
		// Purposely not returning error
		return nil
	}

	return CreateSymlink(expanded, expanded[:len(expanded)-len(distSuffix)])
}

func setupLogging(config *Config) error {
	if config.LogFile == "" {
		return nil
	}

	if err := createDirs(path.Dir(config.LogFile)); err != nil {
		return err
	}

	output, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	syscall.Dup3(int(output.Fd()), int(os.Stdout.Fd()), 0)
	syscall.Dup3(int(output.Fd()), int(os.Stderr.Fd()), 0)

	return nil
}

func setUlimit(cfg *Config) error {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		return err
	}
	if cfg.NoFiles == 0 {
		rLimit.Max = 1000000
	} else {
		rLimit.Max = cfg.NoFiles
	}
	rLimit.Cur = rLimit.Max
	return syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
}

func runOrExec(config *Config, docker string, args ...string) (*exec.Cmd, error) {
	if err := secondPrepare(config, docker, args...); err != nil {
		return nil, err
	}

	cmd := path.Base(docker)
	if config != nil && config.CommandName != "" {
		cmd = config.CommandName
	}

	if cmd == "dockerd" && len(args) > 1 && args[0] == "daemon" {
		args = args[1:]
	}

	return execDocker(config, docker, cmd, args)
}

func LaunchDocker(config *Config, docker string, args ...string) (*exec.Cmd, error) {
	if err := PrepareFs(config); err != nil {
		return nil, err
	}

	return runOrExec(config, docker, args...)
}

func Main() {
	log.InitLogger()
	if os.Getenv("DOCKER_LAUNCH_DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage Example: %s /usr/bin/docker -d -D", os.Args[0])
	}

	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	var config Config
	args = ParseConfig(&config, args...)

	if os.Getenv("DOCKER_LAUNCH_REAP") == "true" {
		config.Fork = true
		config.PidOne = true
	}

	log.Debugf("Launch config %#v", config)

	_, err := LaunchDocker(&config, os.Args[1], args...)
	if err != nil {
		log.Fatal(err)
	}
}
