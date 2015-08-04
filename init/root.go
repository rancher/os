package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/rancher/docker-from-scratch"
	"github.com/rancherio/os/config"
)

func prepareRoot(rootfs string) error {
	usr := path.Join(rootfs, "usr")
	if err := os.Remove(usr); err != nil && !os.IsNotExist(err) {
		log.Errorf("Failed to delete %s, possibly invalid RancherOS state partition: %v", usr, err)
		return err
	}

	return nil
}

func copyMoveRoot(rootfs string) error {
	usrVer := fmt.Sprintf("usr-%s", config.VERSION)
	usr := path.Join(rootfs, usrVer)

	if err := archive.CopyWithTar("/usr", usr); err != nil {
		return err
	}

	if err := dockerlaunch.CreateSymlink(usrVer, path.Join(rootfs, "usr")); err != nil {
		return err
	}

	files, err := ioutil.ReadDir("/")
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := path.Join("/", file.Name())

		if filename == rootfs {
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

func switchRoot(rootfs string) error {
	for _, i := range []string{"/dev", "/sys", "/proc", "/run"} {
		log.Debugf("Moving mount %s to %s", i, path.Join(rootfs, i))
		if err := os.MkdirAll(path.Join(rootfs, i), 0755); err != nil {
			return err
		}
		if err := syscall.Mount(i, path.Join(rootfs, i), "", syscall.MS_MOVE, ""); err != nil {
			return err
		}
	}

	if err := copyMoveRoot(rootfs); err != nil {
		return err
	}

	if err := syscall.Chdir(rootfs); err != nil {
		return err
	}

	if err := syscall.Mount(rootfs, "/", "", syscall.MS_MOVE, ""); err != nil {
		return err
	}

	if err := syscall.Chroot("."); err != nil {
		return err
	}

	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	log.Debugf("Successfully moved to new root at %s", rootfs)
	os.Unsetenv("DOCKER_RAMDISK")

	return nil
}
