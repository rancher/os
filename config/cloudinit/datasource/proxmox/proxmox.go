package proxmox

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"

	"github.com/burmilla/os/config/cloudinit/datasource"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"

	"github.com/docker/docker/pkg/mount"
)

const (
	configDev           = "/dev/sr0"
	configDevMountPoint = "/media/pve-config"
)

type Proxmox struct {
	root                string
	readFile            func(filename string) ([]byte, error)
	lastError           error
	availabilityChanges bool
}

func NewDataSource(root string) *Proxmox {
	return &Proxmox{root, ioutil.ReadFile, nil, true}
}

func (pve *Proxmox) IsAvailable() bool {
	if pve.root == configDevMountPoint {
		pve.lastError = MountConfigDrive()
		if pve.lastError != nil {
			log.Error(pve.lastError)
			pve.availabilityChanges = false
			return false
		}
		defer pve.Finish()
	}

	_, pve.lastError = os.Stat(pve.root)
	return !os.IsNotExist(pve.lastError)
}

func (pve *Proxmox) Finish() error {
	return UnmountConfigDrive()
}

func (pve *Proxmox) String() string {
	if pve.lastError != nil {
		return fmt.Sprintf("%s: %s (lastError: %v)", pve.Type(), pve.root, pve.lastError)
	}
	return fmt.Sprintf("%s: %s", pve.Type(), pve.root)
}

func (pve *Proxmox) AvailabilityChanges() bool {
	return pve.availabilityChanges
}

func (pve *Proxmox) ConfigRoot() string {
	return pve.root
}

func (pve *Proxmox) FetchMetadata() (metadata datasource.Metadata, err error) {
	return datasource.Metadata{}, nil
}

func (pve *Proxmox) FetchUserdata() ([]byte, error) {
	return pve.tryReadFile(path.Join(pve.root, "user-data"))
}

func (pve *Proxmox) Type() string {
	return "proxmox"
}

func (pve *Proxmox) tryReadFile(filename string) ([]byte, error) {
	if pve.root == configDevMountPoint {
		pve.lastError = MountConfigDrive()
		if pve.lastError != nil {
			log.Error(pve.lastError)
			return nil, pve.lastError
		}
		defer pve.Finish()
	}
	log.Debugf("Attempting to read from %q\n", filename)
	data, err := pve.readFile(filename)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		log.Errorf("ERROR read cloud-config file(%s) - err: %q", filename, err)
	}
	return data, err
}

func MountConfigDrive() error {
	if err := os.MkdirAll(configDevMountPoint, 700); err != nil {
		return err
	}

	fsType, err := util.GetFsType(configDev)
	if err != nil {
		return err
	}

	return mount.Mount(configDev, configDevMountPoint, fsType, "ro")
}

func UnmountConfigDrive() error {
	return syscall.Unmount(configDevMountPoint, 0)
}
