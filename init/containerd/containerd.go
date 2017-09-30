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

	composeConfig "github.com/docker/libcompose/config"
	specs "github.com/opencontainers/runtime-spec/specs-go"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"

	"github.com/docker/distribution/reference"

	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/init/prepare"
	"github.com/rancher/os/log"

	"github.com/containerd/containerd/linux/runcopts"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

func LaunchDaemon() error {
	systemInitCmd([]string{})
	return nil
}

type Tree struct {
	runafter  *Tree
	Name      string
	Reason    string
	runbefore *Tree
	parent    *Tree
}

type ServicesOrder struct {
	tree *Tree
	Map  map[string]*Tree
}

func (o ServicesOrder) insert(cfg *config.CloudConfig, s *composeConfig.ServiceConfigV1) ServicesOrder {

	if _, ok := o.Map[s.Name]; ok {
		// lets not add the same service twice
		return o
	}
	e := &Tree{Name: s.Name}
	o.Map[s.Name] = e

	after, _ := s.Labels["io.rancher.os.after"]
	before, _ := s.Labels["io.rancher.os.before"]

	if after != "" {
		e.Reason = e.Reason + "(after)" + after
		a, ok := o.Map[after]
		if !ok {
			s := cfg.Rancher.Services[after]
			o = o.insert(cfg, s)
			a, ok = o.Map[after]
		}
		if a.parent != nil {
			// place the new element where the old one was.
			e.parent = a.parent
			if a.parent.runbefore == a {
				a.parent.runbefore = e
			}
			if a.parent.runafter == a {
				a.parent.runafter = e
			}
		}
		a.parent = e
		e.runafter = a
		if a == o.tree {
			o.tree = e
		}
	}
	if before != "" {
		e.Reason = e.Reason + "(before)" + before
		b, ok := o.Map[before]
		if !ok {
			s := cfg.Rancher.Services[before]
			o = o.insert(cfg, s)
			b, ok = o.Map[before]
		}
		if b.parent != nil {
			// place the new element where the old one was.
			e.parent = b.parent
			if b.parent.runbefore == b {
				b.parent.runbefore = e
			}
			if b.parent.runafter == b {
				b.parent.runafter = e
			}
		}
		b.parent = e
		e.runbefore = b
		if b == o.tree {
			o.tree = e
		}
	}

	if o.tree == nil && e.parent == nil {
		o.tree = e
		return o
	}

	if e.parent == nil && o.tree != e {
		// run last?
		var newParent = o.tree
		for newParent.runafter != nil {
			if newParent.runafter == o.tree {
				newParent.runafter = nil
			}
			newParent = newParent.runafter
		}
		newParent.runafter = e
		e.parent = newParent
	}

	return o
}

func Walk(t *Tree, ch chan *Tree) {
	if t == nil {
		return
	}
	Walk(t.runafter, ch)
	ch <- t
	Walk(t.runbefore, ch)
}

func (o ServicesOrder) Walker() <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		Walk(o.tree, ch)
		close(ch)
	}()
	return ch
}

func RunSet() error {
	cfg := config.LoadConfig()

	// Set the service name
	i := 1
	for name, _ := range cfg.Rancher.Services {
		//		log.Infof("naming (%d): %s", i, name)
		i = i + 1
		s := cfg.Rancher.Services[name]
		s.Name = name
	}
	//order these based on scope labels
	// unfortunantely, its a go hash, so the order (especially of the unconstrained services) will be random
	order := ServicesOrder{Map: make(map[string]*Tree)}
	for name, _ := range cfg.Rancher.Services {
		s := cfg.Rancher.Services[name]
		//		log.Infof("inserting: %s", s.Name)
		order = order.insert(cfg, s)
	}
	log.Infof("services order.")
	ch := order.Walker()
	i = 1
	for {
		t, ok := <-ch
		if !ok {
			break
		}
		log.Infof(" - (%d)%s: %s", i, t.Name, t.Reason)
		i = i + 1
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
		if err := Run(name, ""); err != nil {
			log.Infof("NOTOK: %s (%s)", name, err)
		} else {
			log.Infof("OK   : %s", name)
		}
	}

	return nil
}
func Run(serviceName, bundleDir string) error {
	cfg := config.LoadConfig()

	service := cfg.Rancher.Services[serviceName]

	if service == nil {
		fmt.Printf("Specified serviceName (%s) not found in RancherOS config", serviceName)
		return fmt.Errorf("Specified serviceName (%s) not found in RancherOS config", serviceName)
	}

	if bundleDir == "" {
		// TODO: use the os-config image name to find the base bundle.
		image, err := reference.ParseNamed(service.Image)
		if err != nil {
			bundleDir, _ = os.Getwd()
		} else {
			n := strings.Split(image.Name(), "/")
			name := n[len(n)-1]
			bundleDir = filepath.Join("/containers/services", name)
		}
	}
	if _, err := os.Stat(bundleDir); err != nil && os.IsNotExist(err) {
		fmt.Printf("Bundle Dir (%s) not found", bundleDir)
		return fmt.Errorf("Bundle Dir (%s) not found", bundleDir)
	}

	// TODO: instead of copying a canned spec file, need to generate from the os-config entry
	cannedSpec := filepath.Join("/usr/share/spec/", serviceName+".spec")
	if err := dfs.CopyFileOverwrite(cannedSpec, bundleDir, "config.json", true); err != nil {
		fmt.Printf("Failed to copy %s into bundleDir %s", cannedSpec, bundleDir)
		return fmt.Errorf("Failed to copy %s into bundleDir %s", cannedSpec, bundleDir)
	}

	// TODO: either add a rw layer over the original bundle, or copy it to a new location

	// need to set ourselves as a child subreaper or we cannot wait for runc as reparents to init
	//if err := sys.SetSubreaper(1); err != nil {
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0); err != nil {
		log.Errorf("Cannot set as subreaper: %v", err)
	}

	err := start(serviceName, bundleDir, service)
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
	defaultSocket     = "/run/containerd/containerd.sock"
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
	_ = os.Remove(defaultSocket)

	// start up containerd
	cmd := exec.Command(defaultContainerd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Errorf("cannot start containerd: %s", err)
	}

	// wait for containerd socket to appear
	for {
		_, err := os.Stat(defaultSocket)
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
	client, err := containerd.New(defaultSocket)
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

func start(serviceName, basePath string, service *composeConfig.ServiceConfigV1) error {
	//path := filepath.Join(basePath, serviceName)
	path := basePath

	rootfs := filepath.Join(path, "rootfs")

	if err := prepare.Filesystem(path, service); err != nil {
		return fmt.Errorf("preparing filesystem: %s", err)
	}

	client, err := containerd.New(defaultSocket)
	if err != nil {
		return fmt.Errorf("creating containerd client: %s", err)
	}

	ctx := namespaces.WithNamespace(context.Background(), "default")

	var spec *specs.Spec
	specf, err := os.Open(filepath.Join(path, "config.json"))
	if err != nil {
		return fmt.Errorf("failed to read service spec: %s", err)
	}
	if err := json.NewDecoder(specf).Decode(&spec); err != nil {
		return fmt.Errorf("failed to parse service spec: %s", err)
	}

	spec.Root.Path = rootfs

	/*	if dumpSpec != "" {
			d, err := os.Create(dumpSpec)
			if err != nil {
				return "", 0, "failed to open file for spec dump", err
			}
			enc := json.NewEncoder(d)
			enc.SetIndent("", "    ")
			if err := enc.Encode(&spec); err != nil {
				return "", 0, "failed to write spec dump", err
			}

		}
	*/
	ctr, err := client.NewContainer(ctx,
		serviceName,
		containerd.WithSpec(spec),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %s", err)
	}

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

func WithNoPivotRoot() containerd.NewTaskOpts {
	return func(_ context.Context, _ *containerd.Client, r *containerd.TaskInfo) error {
		r.Options = &runcopts.CreateOptions{
			NoPivotRoot: true,
		}
		return nil
	}
}
