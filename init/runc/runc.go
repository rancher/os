// +build amd64,linux
package runc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	//"github.com/containerd/containerd/sys"
	"golang.org/x/sys/unix"

	composeConfig "github.com/docker/libcompose/config"

	"github.com/docker/distribution/reference"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/init/prepare"
	"github.com/rancher/os/log"
)

// RunSet runs all the services in the list
func RunSet(cfg *config.CloudConfig, serviceSet string, pivotRoot bool) error {
	order := prepare.GetServicesInOrder(cfg, serviceSet)

	log.Infof("Running services.")
	ch := order.Walker()
	for {
		t, ok := <-ch
		if !ok {
			break
		}
		Run(cfg, serviceSet, t.Name, "", pivotRoot)
	}

	return nil
}

// Run can be used to start a service listed in rancher.services, rancher.bootstrap, or rancher.cloud_init_services
func Run(cfg *config.CloudConfig, serviceSet, serviceName, bundleDir string, pivotRoot bool) error {
	service := prepare.GetService(cfg, serviceSet, serviceName)

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
	// https://groups.google.com/a/opencontainers.org/forum/#!topic/dev/ntwTxl9hFp4

	// need to set ourselves as a child subreaper or we cannot wait for runc as reparents to init
	//if err := sys.SetSubreaper(1); err != nil {
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0); err != nil {
		log.Errorf("Cannot set as subreaper: %v", err)
	}

	err := runc(serviceName, bundleDir, pivotRoot, service)
	if err != nil {
		fmt.Printf("Runc error: %s", err)
	} else {
		fmt.Printf("Runc ok")
	}
	return err
}

const (
	runcBinary = "/usr/bin/runc"
)

func runc(serviceName, bundleDir string, noPivot bool, service *composeConfig.ServiceConfigV1) error {
	if err := prepare.Filesystem(bundleDir, service); err != nil {
		return fmt.Errorf("Error preparing %s: %v", serviceName, err)
	}
	tmpdir := "/tmp"
	pidfile := filepath.Join(tmpdir, serviceName)
	args := []string{
		"create", "--bundle", bundleDir, "--pid-file", pidfile,
	}
	if noPivot {
		log.Infof("Starting runc service with --no-pivot")
		args = append(args, "--no-pivot")
	}
	args = append(args, serviceName)
	cmd := exec.Command(runcBinary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error creating %s: %v", serviceName, err)
	}
	pf, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return fmt.Errorf("Cannot read pidfile: %v", err)
	}
	pid, err := strconv.Atoi(string(pf))
	if err != nil {
		return fmt.Errorf("Cannot parse pid from pidfile: %v", err)
	}

	/*if err := prepare.Process(pid); err != nil {
		return fmt.Errorf("Cannot prepare process: %v", err)
	}*/

	waitFor := make(chan *os.ProcessState)
	go func() {
		// never errors in Unix
		p, _ := os.FindProcess(pid)
		state, err := p.Wait()
		if err != nil {
			log.Errorf("Process wait error: %v", err)
		}
		waitFor <- state
	}()

	cmd = exec.Command(runcBinary, "start", serviceName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error starting %s: %v", serviceName, err)
	}

	_ = <-waitFor

	cleanup(bundleDir)
	_ = os.Remove(pidfile)

	//_ = os.RemoveAll(tmpdir)

	return nil
}

// bind mount a namespace file
func bindNS(ns string, path string, pid int) error {
	if path == "" {
		return nil
	}
	// the path and file need to exist for the bind to succeed, so try to create
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Cannot create leading directories %s for bind mount destination: %v", dir, err)
	}
	fi, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Cannot create a mount point for namespace bind at %s: %v", path, err)
	}
	if err := fi.Close(); err != nil {
		return err
	}
	if err := unix.Mount(fmt.Sprintf("/proc/%d/ns/%s", pid, ns), path, "", unix.MS_BIND, ""); err != nil {
		return fmt.Errorf("Failed to bind %s namespace at %s: %v", ns, path, err)
	}
	return nil
}

// cleanup functions are best efforts only, mainly for rw onboot containers
func cleanup(path string) {
	// remove the root mount
	rootfs := filepath.Join(path, "rootfs")
	_ = unix.Unmount(rootfs, 0)
}
