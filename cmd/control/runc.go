package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/codegangsta/cli"

	composeConfig "github.com/docker/libcompose/config"

	"github.com/docker/docker/reference"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
)

func runcCommand() cli.Command {
	var pivot cli.Flag
	if util.RootFsIsNotReal() {
		pivot = cli.BoolFlag{
			Name:  "pivot-root",
			Usage: "pivot-root (defaulted to false due to tmmpfs/ramfs)",
		}
	} else {
		pivot = cli.BoolTFlag{
			Name:  "pivot-root",
			Usage: "pivot-root (defaulted to true)",
		}
	}

	return cli.Command{
		Name:   "runc",
		Usage:  "create, prepare and run using runc",
		Action: runcAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "bundle, b",
				Usage: "path to the root of the bundle dir",
			},
			pivot,
			// TODO: add a --delete ?
		},
	}
}
func runcAction(c *cli.Context) error {
	fmt.Print("Runc start\n")
	serviceName := c.Args().Get(0)
	if serviceName == "" {
		fmt.Print("Please specify the service name (needs to be in the os-config)")
		return fmt.Errorf("Please specify the service name (needs to be in the os-config)")
	}
	bundleDir := c.String("bundle")
	pivotRoot := c.Bool("pivot-root")
	return Runc(serviceName, bundleDir, pivotRoot)
}

func Runc(serviceName, bundleDir string, pivotRoot bool) error {
	cfg := config.LoadConfig()
	service := cfg.Rancher.Services[serviceName]
	if service == nil {
		// maybe its bootstrap or cloud_init_services
		service = cfg.Rancher.BootstrapContainers[serviceName]
		if service == nil {
			service = cfg.Rancher.CloudInitServices[serviceName]
		}
	}
	if service == nil {
		fmt.Print("Specified serviceName (%s) not found in RancherOS config", serviceName)
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
		fmt.Print("Bundle Dir (%s) not found", bundleDir)
		return fmt.Errorf("Bundle Dir (%s) not found", bundleDir)
	}

	// TODO: instead of copying a canned spec file, need to generate from the os-config entry
	cannedSpec := filepath.Join("/usr/share/spec/", serviceName+".spec")
	if err := dfs.CopyFileOverwrite(cannedSpec, bundleDir, "config.json", true); err != nil {
		fmt.Print("Failed to copy %s into bundleDir %s", cannedSpec, bundleDir)
		return fmt.Errorf("Failed to copy %s into bundleDir %s", cannedSpec, bundleDir)
	}

	// TODO: either add a rw layer over the original bundle, or copy it to a new location

	err := runc(serviceName, bundleDir, !pivotRoot, service)
	if err != nil {
		fmt.Print("Runc error: %s\n", err)
	} else {
		fmt.Printf("Runc ok\n")
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
			log.Printf("Process wait error: %v", err)
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
	// execute the runtime config that should be done up front
	// we execute Mounts before Mkdir so you can make a directory under a mount
	// but we do mkdir of the destination path in case missing
	for _, volume := range service.Volumes {
		v := strings.Split(volume, ":")
		source := v[0]
		destination := filepath.Join(path, "rootfs", v[1])
		//options := v[2]
		mountType := "bind"

		fmt.Printf("Volume(%s)\n", v)
		fmt.Printf("  dest: %s\n", destination)
		fmt.Printf("  src: %s\n", source)

		s, err := os.Stat(source)
		mkdir := destination
		destFile := ""
		switch {
		case err != nil:
			fmt.Printf("Error stating %s: %s\n", source, err)
			//			mkdir = ""
			// This is potentially flawed - we might want both to come into existence
			mkdir = filepath.Dir(destination)
			destFile = destination
		case s.IsDir():
			fmt.Printf("MkdirAll(%s)\n", destination)
		default:
			fmt.Printf("stating %s: not a Dir: %s\n", source, s.Mode())
			mkdir = filepath.Dir(destination)
			destFile = destination
		}
		if mkdir != "" {
			fmt.Printf("MkdirAll(%s)\n", mkdir)

			const mode os.FileMode = 0755
			err := os.MkdirAll(mkdir, mode)
			if err != nil {
				fmt.Errorf("Cannot create directory for mount destination %s: %v\n", mkdir, err)
			}
		}
		// if the source is a file, then create the destination file too
		if destFile != "" {
			f, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE, s.Mode())
			if err != nil {
				fmt.Errorf("Cannot create file for mount destination %s: %v\n", destFile, err)
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
				fmt.Printf("Error stating %s: %s\n", destination, err)
			case d.IsDir():
				fmt.Printf("MkdirAll(%s)\n", destination)
			default:
				fmt.Printf("stating %s: not a Dir: %s\n", destination, d.Mode())
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
