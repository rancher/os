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

	"github.com/docker/docker/reference"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/log"
)

// RunSet runs all the services in the list
// TODO: extract from RunC once we have containerd
func RunSet(serviceSet string, pivotRoot bool) error {
	set := getServiceSet(serviceSet)
	//TODO: need to order these based on scope labels
	for name, _ := range set {
		Run(name, "", pivotRoot)
	}

	return nil
}

func getServiceSet(name string) map[string]*composeConfig.ServiceConfigV1 {
	cfg := config.LoadConfig()
	var set map[string]*composeConfig.ServiceConfigV1
	switch name {
	case "services":
		set = cfg.Rancher.Services
	case "bootstrap":
		set = cfg.Rancher.BootstrapContainers
	case "cloud_init_services":
		set = cfg.Rancher.CloudInitServices
	case "recovery_services":
		set = cfg.Rancher.RecoveryServices
	}
	return set
}

func getService(name string) *composeConfig.ServiceConfigV1 {
	cfg := config.LoadConfig()

	switch {
	case cfg.Rancher.Services[name] != nil:
		return cfg.Rancher.Services[name]
	case cfg.Rancher.BootstrapContainers[name] != nil:
		return cfg.Rancher.BootstrapContainers[name]
	case cfg.Rancher.CloudInitServices[name] != nil:
		return cfg.Rancher.CloudInitServices[name]
	case cfg.Rancher.RecoveryServices[name] != nil:
		return cfg.Rancher.RecoveryServices[name]
	}

	return nil
}

// Run can be used to start a service listed in rancher.services, rancher.bootstrap, or rancher.cloud_init_services
func Run(serviceName, bundleDir string, pivotRoot bool) error {
	service := getService(serviceName)

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
	if err := prepareFilesystem(bundleDir, service); err != nil {
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

	/*if err := prepareProcess(pid); err != nil {
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

// prepareFilesystem sets up the mounts, before the container is created
func prepareFilesystem(path string, service *composeConfig.ServiceConfigV1) error {

	// TODO: work out why these dirs are needed (by console), and not on the host fs by default
	const mode os.FileMode = 0755
	os.MkdirAll("/opt", mode)
	os.MkdirAll("/var/lib/rancher/cache", mode)

	// execute the runtime config that should be done up front
	// we execute Mounts before Mkdir so you can make a directory under a mount
	// but we do mkdir of the destination path in case missing
	for _, volume := range service.Volumes {
		v := strings.Split(volume, ":")
		source := v[0]
		destination := filepath.Join(path, "rootfs", v[1])
		//options := v[2]
		mountType := "bind"

		//log.Infof("Volume(%s)", v)
		//log.Infof("  dest: %s", destination)
		//log.Infof("  src: %s", source)

		s, err := os.Stat(source)
		mkdir := destination
		destFile := ""
		switch {
		case err != nil:
			log.Errorf("Error stating %s: %s", source, err)
			//			mkdir = ""
			// This is potentially flawed - we might want both to come into existence
			mkdir = filepath.Dir(destination)
			destFile = destination
		case s.IsDir():
		default:
			log.Infof("stating %s: not a Dir: %s", source, s.Mode())
			mkdir = filepath.Dir(destination)
			destFile = destination
		}
		if mkdir != "" {
			log.Infof("MkdirAll(%s)", mkdir)

			const mode os.FileMode = 0755
			err := os.MkdirAll(mkdir, mode)
			if err != nil {
				log.Errorf("Cannot create directory for mount destination %s: %v", mkdir, err)
			}
		}
		// if the source is a file, then create the destination file too
		if destFile != "" {
			f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE, s.Mode())
			if err != nil {
				log.Errorf("Cannot create file for mount destination %s: %v", destFile, err)
			}
			f.Close()
		}

		// also mkdir upper and work directories on overlay
		/*		for _, o := range mount.Options {
					eq := strings.SplitN(o, "=", 2)
					if len(eq) == 2 && (eq[0] == "upperdir" || eq[0] == "workdir") {
						err := os.MkdirAll(eq[1], mode)
						if err != nil {
							return fmt.Errorf("Cannot create directory for overlay %s=%s: %v", eq[0], eq[1], err)
						}
					}
				}
				opts, data := parseMountOptions(mount.Options)
		*/
		opts := unix.MS_BIND | unix.MS_REC | unix.MS_PRIVATE
		data := ""
		if err := unix.Mount(source, destination, mountType, uintptr(opts), data); err != nil {
			d, err := os.Stat(destination)
			switch {
			case err != nil:
				log.Errorf("Error stating %s: %s", destination, err)
			case d.IsDir():
				log.Infof("MkdirAll(%s)", destination)
			default:
				log.Infof("stating %s: not a Dir: %s", destination, d.Mode())
			}

			return fmt.Errorf("Failed to mount %s to %s : %v", source, destination, err)
		}
	}

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

// prepareProcess sets up anything that needs to be done after the container process is created, but before it runs
// for example networking
/*func prepareProcess(pid int) error {
	binds := []struct {
		ns   string
		path string
	}{
		{"cgroup", runtime.BindNS.Cgroup},
		{"ipc", runtime.BindNS.Ipc},
		{"mnt", runtime.BindNS.Mnt},
		{"net", runtime.BindNS.Net},
		{"pid", runtime.BindNS.Pid},
		{"user", runtime.BindNS.User},
		{"uts", runtime.BindNS.Uts},
	}

	for _, b := range binds {
		if err := bindNS(b.ns, b.path, pid); err != nil {
			return err
		}
	}

	return nil
}
*/
// cleanup functions are best efforts only, mainly for rw onboot containers
func cleanup(path string) {
	// remove the root mount
	rootfs := filepath.Join(path, "rootfs")
	_ = unix.Unmount(rootfs, 0)
}
