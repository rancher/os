package containerd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/rancher/os/config"

	composeConfig "github.com/docker/libcompose/config"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"

	"github.com/rancher/os/init/prepare"
	"github.com/rancher/os/log"

	"github.com/docker/distribution/reference"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/linux/runcopts"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func LaunchDaemon() error {
	systemInitCmd([]string{})
	return nil
}

func RunSet(cfg *config.CloudConfig, serviceSet string) error {
	order := prepare.GetServicesInOrder(cfg, serviceSet)

	log.Infof("Services order")
	ch := order.Walker()
	for {
		t, ok := <-ch
		if !ok {
			break
		}
		name := t.Name

		log.Infof("- %s", name)
	}

	log.Infof("Running services.")
	ch = order.Walker()
	for {
		t, ok := <-ch
		if !ok {
			break
		}
		name := t.Name

		log.Infof("STARTING: %s", name)
		if err := Run(cfg, serviceSet, name, ""); err != nil {
			log.Infof("NOTOK: %s (%s)", name, err)
		} else {
			log.Infof("OK   : %s", name)
		}
	}

	return nil
}
func Run(cfg *config.CloudConfig, serviceSet, serviceName, bundleDir string) error {
	service := prepare.GetService(cfg, serviceSet, serviceName)

	if service == nil {
		fmt.Printf("Specified serviceName (%s) not found in RancherOS config", serviceName)
		return fmt.Errorf("Specified serviceName (%s) not found in RancherOS config", serviceName)
	}

	// need to set ourselves as a child subreaper or we cannot wait for runc as reparents to init
	//if err := sys.SetSubreaper(1); err != nil {
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0); err != nil {
		log.Errorf("Cannot set as subreaper: %v", err)
	}

	// Where the images, and then the running overlay fs's live (for now)
	basePath := "/containers/services"
	specFile := filepath.Join("/usr/share/spec/", serviceName+".spec")
	if _, err := os.Stat(specFile); err != nil {
		// presumably this is a user-customisation, so we'll just "try"
		specFile = filepath.Join("/usr/share/spec/console.spec")
	}
	err := start(cfg, serviceSet, basePath, specFile, serviceName, service)
	if err != nil {
		log.Infof("Runc error: %s", err)
	} else {
		log.Infof("Runc ok")
	}
	return err
}

// yup, exact copy from linuxkit
func cleanupTask(ctx context.Context, ctr containerd.Container) error {
	task, err := ctr.Task(ctx, nil)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return errors.Wrap(err, "getting task")
	}

	deleteErr := make(chan error, 1)
	deleteCtx, deleteCancel := context.WithCancel(ctx)
	defer deleteCancel()

	go func(ctx context.Context, ch chan error) {
		_, err := task.Delete(ctx)
		if err != nil {
			ch <- errors.Wrap(err, "killing task")
		}
		ch <- nil
	}(deleteCtx, deleteErr)

	sig := syscall.SIGKILL
	if err := task.Kill(ctx, sig); err != nil && !errdefs.IsNotFound(err) {
		return errors.Wrapf(err, "killing task with %q", sig)
	}

	select {
	case err := <-deleteErr:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

const (
	//DefaultContainerdSocket = "/run/containerd/containerd.sock"
	defaultPath       = "/containers/services"
	defaultContainerd = "/usr/bin/containerd"
	installPath       = "/usr/bin/service"
	onbootPath        = "/containers/onboot"
	shutdownPath      = "/containers/onshutdown"
)

func systemInitCmd(args []string) {
	invoked := filepath.Base(os.Args[0])
	flags := flag.NewFlagSet("system-init", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Printf("USAGE: %s system-init\n\n", invoked)
		fmt.Printf("Options:\n")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		log.Fatal("Unable to parse args")
	}
	args = flags.Args()

	if len(args) != 0 {
		fmt.Println("Unexpected argument")
		flags.Usage()
		os.Exit(1)
	}

	// remove (unlikely) old containerd socket
	_ = os.Remove(config.DefaultContainerdSocket)

	// start up containerd
	cmd := exec.Command(defaultContainerd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Errorf("cannot start containerd: %s", err)
	}

	// wait for containerd socket to appear
	for {
		_, err := os.Stat(config.DefaultContainerdSocket)
		if err == nil {
			break
		}
		err = cmd.Process.Signal(syscall.Signal(0))
		if err != nil {
			// process not there, wait() to find error
			err = cmd.Wait()
			log.Errorf("containerd process exited: %s", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// connect to containerd
	client, err := containerd.New(config.DefaultContainerdSocket)
	if err != nil {
		log.Errorf("creating containerd client: %s", err)
	}

	ctx := namespaces.WithNamespace(context.Background(), "default")

	ctrs, err := client.Containers(ctx)
	if err != nil {
		log.Errorf("listing containers: %s", err)
	}

	// Clean up any old containers
	// None of the errors in this loop are fatal since we want to
	// keep trying.
	for _, ctr := range ctrs {
		log.Infof("Cleaning up stale service: %q", ctr.ID())
		log := log.WithFields(log.Fields{
			"service": ctr.ID(),
		})

		if err := cleanupTask(ctx, ctr); err != nil {
			log.WithError(err).Error("cleaning up task")
		}

		if err := ctr.Delete(ctx); err != nil {
			log.WithError(err).Error("deleting container")
		}
	}

	// Start up containers
	//	files, err := ioutil.ReadDir(defaultPath)
	//	// just skip if there is an error, eg no such path
	//	if err != nil {
	//		return
	//	}
	//	for _, file := range files {
	//		if id, pid, msg, err := start(file.Name(), *sock, *path, ""); err != nil {
	//			log.WithError(err).Error(msg)
	//		} else {
	//			log.Debugf("Started %s pid %d", id, pid)
	//		}
	//	}
}

type cio struct {
	config containerd.IOConfig
}

func (c *cio) Config() containerd.IOConfig {
	return c.config
}

func (c *cio) Cancel() {
}

func (c *cio) Wait() {
}

func (c *cio) Close() error {
	return nil
}

func start(cfg *config.CloudConfig, serviceSet, basePath, specFile, serviceName string, service *composeConfig.ServiceConfigV1) error {
	path := filepath.Join(basePath, serviceName)

	image, err := reference.ParseNamed(service.Image)
	if err != nil {
		return fmt.Errorf("failed to parse image name from %s: %s", service.Image, err)
	}
	n := strings.Split(image.Name(), "/")
	imageName := n[len(n)-1]

	rootfs := filepath.Join(basePath, imageName, "rootfs")

	if err := prepare.Filesystem(path, service); err != nil {
		return fmt.Errorf("preparing filesystem: %s", err)
	}

	client, err := containerd.New(config.DefaultContainerdSocket)
	if err != nil {
		return fmt.Errorf("creating containerd client: %s", err)
	}

	ctx := namespaces.WithNamespace(context.Background(), "default")

	var spec *specs.Spec
	specf, err := os.Open(specFile)
	if err != nil {
		return fmt.Errorf("failed to read service spec: %s", err)
	}
	if err := json.NewDecoder(specf).Decode(&spec); err != nil {
		return fmt.Errorf("failed to parse service spec: %s", err)
	}

	// TODO: this means we're not using containerd images, and can't use its snapshotting
	//spec.Root.Path = rootfs

	// the overlay dirs need to exist...
	workDir := filepath.Join("/containers", "work", serviceName)
	if err = os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("mkdirall : %s/work", err)
	}
	rwDir := filepath.Join("/containers", "rw", serviceName)
	if err = os.MkdirAll(rwDir, 0755); err != nil {
		return fmt.Errorf("mkdirall : %s/rw", err)
	}

	//{
	// attempt to build a spec file from scratch, using the service struct
	ctr, err := client.NewContainer(ctx,
		serviceName,
		//containerd.WithNewSnapshot(serviceName+"-snapshot", image),
		containerd.WithNewSpec(
			withService(cfg, serviceSet, service),
			removeRunTmpfsMount(),
			withDevicesFromSpec(spec),
			withOverlay(rootfs, rwDir, workDir),
			dumpSpec(serviceName+".new"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %s", err)
	}

	//}
	/*ctr, err := client.NewContainer(ctx,
		serviceName,
		//containerd.WithNewSnapshot(serviceName+"-snapshot", image),
		containerd.WithSpec(
			spec,
			withOverlay(rootfs, rwDir, workDir),
			dumpSpec(serviceName+".actual"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %s", err)
	}*/

	io := func(id string) (containerd.IO, error) {
		logfile := filepath.Join("/var/log", serviceName+".log")
		// We just need this to exist.
		if err := ioutil.WriteFile(logfile, []byte{}, 0600); err != nil {
			// if we cannot write to log, discard output
			logfile = "/dev/null"
		}
		return &cio{
			containerd.IOConfig{
				Stdin:    "/dev/null",
				Stdout:   logfile,
				Stderr:   logfile,
				Terminal: false,
			},
		}, nil
	}

	//task, err := ctr.NewTask(ctx, io)
	task, err := ctr.NewTask(ctx, io, WithNoPivotRoot())

	if err != nil {
		// Don't bother to destroy the container here.
		return fmt.Errorf("failed to create task: %s", err)
	}

	//if err := prepare.Process(int(task.Pid()), runtimeConfig); err != nil {
	//	return "", 0, "preparing process", err
	//}

	if err := task.Start(ctx); err != nil {
		// Don't destroy the container here so it can be inspected for debugging.
		return fmt.Errorf("failed to start task: %s", err)
	}

	return nil
}

// Yes, we're not using a tmpfs for /run - we're sharing the host one for system-services
func removeRunTmpfsMount() containerd.SpecOpts {
	return func(_ context.Context, _ *containerd.Client, _ *containers.Container, s *specs.Spec) error {
		for i, mount := range s.Mounts {
			if mount.Destination == "/run" && mount.Type == "tmpfs" {
				   s.Mounts[i] = s.Mounts[len(s.Mounts)-1]
				   s.Mounts = s.Mounts[:len(s.Mounts)-1]
				   return nil
			}
		}
		return nil
	}
}

func dumpSpec(serviceName string) containerd.SpecOpts {
	return func(_ context.Context, _ *containerd.Client, _ *containers.Container, s *specs.Spec) error {
		d, err := os.Create(filepath.Join("/containers/", serviceName))
		if err != nil {
			log.Errorf("failed to open file for spec dump: %s", err)
			return fmt.Errorf("failed to open file for spec dump: %s", err)
		}
		enc := json.NewEncoder(d)
		enc.SetIndent("", "    ")
		if err := enc.Encode(s); err != nil {
			log.Errorf("failed to write file for spec dump: %s", err)
			return fmt.Errorf("failed to write file for spec dump: %s", err)
		}
		log.Infof("spec dump ok: %s", serviceName)

		return nil
	}
}

func addVolumes(cfg *config.CloudConfig, serviceSet string, service *composeConfig.ServiceConfigV1, s *specs.Spec) {
	for _, fromService := range service.VolumesFrom {
		from := prepare.GetService(cfg, serviceSet, fromService)
		addVolumes(cfg, serviceSet, from, s)
	}

	//{
	//	"destination": "/etc/docker",
	//	"type": "bind",
	//	"source": "/etc/docker",
	//	"options": [
	//	"rw",
	//		"rbind",
	//		"rprivate"
	//	]
	//},
	for _, vol := range service.Volumes {
		mode := "rw"
		v := strings.Split(vol, ":")
		if len(v) > 2 {
			mode = v[2]
		}
		s.Mounts = append([]specs.Mount{
			specs.Mount{
				Source:      v[0],
				Destination: v[1],
				Type:        "bind",
				Options: []string{
					mode,
					"rbind",
					"rprivate",
				},
			},
		}, s.Mounts...)
	}

}

var privilegedCaps = []string{
	"CAP_CHOWN",
	"CAP_DAC_OVERRIDE",
	"CAP_DAC_READ_SEARCH",
	"CAP_FOWNER",
	"CAP_FSETID",
	"CAP_KILL",
	"CAP_SETGID",
	"CAP_SETUID",
	"CAP_SETPCAP",
	"CAP_LINUX_IMMUTABLE",
	"CAP_NET_BIND_SERVICE",
	"CAP_NET_BROADCAST",
	"CAP_NET_ADMIN",
	"CAP_NET_RAW",
	"CAP_IPC_LOCK",
	"CAP_IPC_OWNER",
	"CAP_SYS_MODULE",
	"CAP_SYS_RAWIO",
	"CAP_SYS_CHROOT",
	"CAP_SYS_PTRACE",
	"CAP_SYS_PACCT",
	"CAP_SYS_ADMIN",
	"CAP_SYS_BOOT",
	"CAP_SYS_NICE",
	"CAP_SYS_RESOURCE",
	"CAP_SYS_TIME",
	"CAP_SYS_TTY_CONFIG",
	"CAP_MKNOD",
	"CAP_LEASE",
	"CAP_AUDIT_WRITE",
	"CAP_AUDIT_CONTROL",
	"CAP_SETFCAP",
	"CAP_MAC_OVERRIDE",
	"CAP_MAC_ADMIN",
	"CAP_SYSLOG",
	"CAP_WAKE_ALARM",
	"CAP_BLOCK_SUSPEND",
	"CAP_AUDIT_READ",
}

func withService(cfg *config.CloudConfig, serviceSet string, service *composeConfig.ServiceConfigV1) containerd.SpecOpts {
	return func(_ context.Context, _ *containerd.Client, _ *containers.Container, s *specs.Spec) error {
		// TODO: need to get the entrypoint info from the docker image :/
		if service.Name == "docker" {
			s.Process.Args = service.Command
		} else if service.Name == "logrotate" || service.Name == "syslog" {
			s.Process.Args = []string{"/usr/bin/entrypoint.sh"}
			s.Process.Args = append(s.Process.Args, service.Command...)
		} else {
			s.Process.Args = []string{"/usr/bin/ros", "entrypoint"}
			s.Process.Args = append(s.Process.Args, service.Command...)
		}

		// allow sudo
		if service.Name == "console" {
			s.Process.NoNewPrivileges = false
		}

		// CAPs
		s.Process.Capabilities.Ambient = privilegedCaps
		s.Process.Capabilities.Bounding = privilegedCaps
		s.Process.Capabilities.Effective = privilegedCaps
		s.Process.Capabilities.Inheritable = privilegedCaps
		s.Process.Capabilities.Permitted = privilegedCaps

		// Volumes and VolumesFrom
		addVolumes(cfg, serviceSet, service, s)

		// Service namespaces
		// TODO: set as per service cfg
		s.Linux.Namespaces = []specs.LinuxNamespace{
			specs.LinuxNamespace{Type: "ipc"},
			specs.LinuxNamespace{Type: "mount"},
			specs.LinuxNamespace{Type: "uts"},
		}
		if service.Name != "console" &&
			service.Name != "docker" {
			s.Linux.Namespaces = append(s.Linux.Namespaces, specs.LinuxNamespace{Type: "pid"})
		}

		return nil
	}
}

func withOverlay(rootfs, rwDir, workDir string) containerd.SpecOpts {
	//{
	//	"destination" : "/",
	//	"options" : [
	//	"upperdir=/containers/services/ntp/rw",
	//		"lowerdir=/containers/services/image/rootfs",
	//		"workdir=/containers/services/ntp/work"
	//	],
	//	"type" : "overlay",
	//	"source" : "overlay"
	//},
	return func(_ context.Context, _ *containerd.Client, _ *containers.Container, s *specs.Spec) error {
		s.Mounts = append([]specs.Mount{
			specs.Mount{
				Destination: "/",
				Options: []string{
					"upperdir=" + rwDir,
					"lowerdir=" + rootfs,
					"workdir=" + workDir,
				},
				Type:   "overlay",
				Source: "overlay",
			},
		}, s.Mounts...)
		return nil
	}
}

func withDevicesFromSpec(fromSpec *specs.Spec) containerd.SpecOpts {
	return func(_ context.Context, _ *containerd.Client, _ *containers.Container, s *specs.Spec) error {
		s.Linux.Resources.Devices = fromSpec.Linux.Resources.Devices
		s.Linux.Devices = fromSpec.Linux.Devices
		return nil
	}
}

func WithNoPivotRoot() containerd.NewTaskOpts {
	return func(_ context.Context, _ *containerd.Client, r *containerd.TaskInfo) error {
		r.Options = &runcopts.CreateOptions{
			NoPivotRoot: true,
		}
		return nil
	}
}
