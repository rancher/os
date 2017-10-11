package prepare

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	composeConfig "github.com/docker/libcompose/config"

	"github.com/rancher/os/log"

	"golang.org/x/sys/unix"
)

// prepareFilesystem sets up the mounts, before the container is created
func Filesystem(path string, service *composeConfig.ServiceConfigV1) error {

	// TODO: work out why these dirs are needed (by console), and not on the host fs by default
	const mode os.FileMode = 0755
	os.MkdirAll("/opt", mode)
	os.MkdirAll("/var/lib/rancher/cache", mode)
	os.MkdirAll("/var/lib/kubelet", mode)

	return nil

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
			log.Errorf("Error stating (1) %s: %s", source, err)
			//			mkdir = ""
			// This is potentially flawed - we might want both to come into existence
			mkdir = filepath.Dir(destination)
			destFile = destination
		case s.IsDir():
		default:
			log.Infof("stating (1) %s: not a Dir: %s", source, s.Mode())
			mkdir = filepath.Dir(destination)
			destFile = destination
		}
		if mkdir != "" {
			log.Infof("MkdirAll (1) (%s)", mkdir)

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
