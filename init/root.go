package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/rancher/os/config"
	"github.com/rancher/os/dfs"
)

func cleanupTarget(rootfs, targetUsr, usr, usrVer, tmpDir string) (bool, error) {
	log.Debugf("Deleting %s", targetUsr)
	if err := os.Remove(targetUsr); err != nil && !os.IsNotExist(err) {
		log.Errorf("Failed to delete %s, possibly invalid RancherOS state partition: %v", targetUsr, err)
		return false, err
	}

	if err := dfs.CreateSymlink(usrVer, path.Join(rootfs, "usr")); err != nil {
		return false, err
	}

	log.Debugf("Deleting %s", tmpDir)
	if err := os.RemoveAll(tmpDir); err != nil {
		// Don't care if this fails
		log.Errorf("Failed to cleanup temp directory %s: %v", tmpDir, err)
	}

	if _, err := os.Stat(usr); err == nil {
		return false, nil
	}

	return true, nil
}

func copyMoveRoot(rootfs string, rmUsr bool) error {
	usrVer := fmt.Sprintf("usr-%s", config.VERSION)
	usr := path.Join(rootfs, usrVer)
	targetUsr := path.Join(rootfs, "usr")
	tmpDir := path.Join(rootfs, "tmp")

	if rmUsr {
		log.Warnf("Development setup. Removing old usr: %s", usr)
		if err := os.RemoveAll(usr); err != nil {
			// Don't care if this fails
			log.Errorf("Failed to remove %s: %v", usr, err)
		}
	}

	if cont, err := cleanupTarget(rootfs, targetUsr, usr, usrVer, tmpDir); !cont {
		return err
	}

	log.Debugf("Creating temp dir directory %s", tmpDir)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}

	usrVerTmp, err := ioutil.TempDir(tmpDir, usrVer)
	if err != nil {
		return err
	}

	log.Debugf("Copying to temp dir %s", usrVerTmp)

	if err := archive.CopyWithTar("/usr", usrVerTmp); err != nil {
		return err
	}

	log.Debugf("Renaming %s => %s", usrVerTmp, usr)
	if err := os.Rename(usrVerTmp, usr); err != nil {
		return err
	}

	files, err := ioutil.ReadDir("/")
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := path.Join("/", file.Name())

		if filename == rootfs || strings.HasPrefix(rootfs, filename+"/") {
			log.Debugf("Skipping Deleting %s", filename)
			continue
		}

		log.Debugf("Deleting %s", filename)
		if err := os.RemoveAll(filename); err != nil {
			return err
		}
	}

	return nil
}

func switchRoot(rootfs, subdir string, rmUsr bool) error {
	if err := syscall.Unmount(config.OEM, 0); err != nil {
		log.Debugf("Not umounting OEM: %v", err)
	}

	if subdir != "" {
		fullRootfs := path.Join(rootfs, subdir)
		if _, err := os.Stat(fullRootfs); os.IsNotExist(err) {
			if err := os.MkdirAll(fullRootfs, 0755); err != nil {
				log.Errorf("Failed to create directory %s: %v", fullRootfs, err)
				return err
			}
		}

		log.Debugf("Bind mounting mount %s to %s", fullRootfs, rootfs)
		if err := syscall.Mount(fullRootfs, rootfs, "", syscall.MS_BIND, ""); err != nil {
			log.Errorf("Failed to bind mount subdir for %s: %v", fullRootfs, err)
			return err
		}
	}

	for _, i := range []string{"/dev", "/sys", "/proc", "/run"} {
		log.Debugf("Moving mount %s to %s", i, path.Join(rootfs, i))
		if err := os.MkdirAll(path.Join(rootfs, i), 0755); err != nil {
			return err
		}
		if err := syscall.Mount(i, path.Join(rootfs, i), "", syscall.MS_MOVE, ""); err != nil {
			return err
		}
	}

	if err := copyMoveRoot(rootfs, rmUsr); err != nil {
		return err
	}

	log.Debugf("chdir %s", rootfs)
	if err := syscall.Chdir(rootfs); err != nil {
		return err
	}

	log.Debugf("mount MS_MOVE %s", rootfs)
	if err := syscall.Mount(rootfs, "/", "", syscall.MS_MOVE, ""); err != nil {
		return err
	}

	log.Debug("chroot .")
	if err := syscall.Chroot("."); err != nil {
		return err
	}

	log.Debug("chdir /")
	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	log.Debugf("Successfully moved to new root at %s", path.Join(rootfs, subdir))
	os.Unsetenv("DOCKER_RAMDISK")

	return nil
}
