// +build !windows

package daemon

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/execdriver"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/docker/runconfig"
	"github.com/docker/docker/volume"
	"github.com/docker/docker/volume/drivers"
	"github.com/docker/docker/volume/local"
	"github.com/opencontainers/runc/libcontainer/label"
)

// copyOwnership copies the permissions and uid:gid of the source file
// into the destination file
func copyOwnership(source, destination string) error {
	stat, err := system.Stat(source)
	if err != nil {
		return err
	}

	if err := os.Chown(destination, int(stat.Uid()), int(stat.Gid())); err != nil {
		return err
	}

	return os.Chmod(destination, os.FileMode(stat.Mode()))
}

func (container *Container) setupMounts() ([]execdriver.Mount, error) {
	var mounts []execdriver.Mount
	for _, m := range container.MountPoints {
		path, err := m.Setup()
		if err != nil {
			return nil, err
		}
		if !container.trySetNetworkMount(m.Destination, path) {
			mounts = append(mounts, execdriver.Mount{
				Source:      path,
				Destination: m.Destination,
				Writable:    m.RW,
			})
		}
	}

	mounts = sortMounts(mounts)
	return append(mounts, container.networkMounts()...), nil
}

func parseBindMount(spec string, mountLabel string, config *runconfig.Config) (*mountPoint, error) {
	bind := &mountPoint{
		RW: true,
	}
	arr := strings.Split(spec, ":")

	switch len(arr) {
	case 2:
		bind.Destination = arr[1]
	case 3:
		bind.Destination = arr[1]
		mode := arr[2]
		isValid, isRw := volume.ValidateMountMode(mode)
		if !isValid {
			return nil, fmt.Errorf("invalid mode for volumes-from: %s", mode)
		}
		bind.RW = isRw
		// Mode field is used by SELinux to decide whether to apply label
		bind.Mode = mode
	default:
		return nil, fmt.Errorf("Invalid volume specification: %s", spec)
	}

	name, source, err := parseVolumeSource(arr[0])
	if err != nil {
		return nil, err
	}

	if len(source) == 0 {
		bind.Driver = config.VolumeDriver
		if len(bind.Driver) == 0 {
			bind.Driver = volume.DefaultDriverName
		}
	} else {
		bind.Source = filepath.Clean(source)
	}

	bind.Name = name
	bind.Destination = filepath.Clean(bind.Destination)
	return bind, nil
}

func sortMounts(m []execdriver.Mount) []execdriver.Mount {
	sort.Sort(mounts(m))
	return m
}

type mounts []execdriver.Mount

func (m mounts) Len() int {
	return len(m)
}

func (m mounts) Less(i, j int) bool {
	return m.parts(i) < m.parts(j)
}

func (m mounts) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m mounts) parts(i int) int {
	return len(strings.Split(filepath.Clean(m[i].Destination), string(os.PathSeparator)))
}

// migrateVolume links the contents of a volume created pre Docker 1.7
// into the location expected by the local driver.
// It creates a symlink from DOCKER_ROOT/vfs/dir/VOLUME_ID to DOCKER_ROOT/volumes/VOLUME_ID/_container_data.
// It preserves the volume json configuration generated pre Docker 1.7 to be able to
// downgrade from Docker 1.7 to Docker 1.6 without losing volume compatibility.
func migrateVolume(id, vfs string) error {
	l, err := getVolumeDriver(volume.DefaultDriverName)
	if err != nil {
		return err
	}

	newDataPath := l.(*local.Root).DataPath(id)
	fi, err := os.Stat(newDataPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if fi != nil && fi.IsDir() {
		return nil
	}

	return os.Symlink(vfs, newDataPath)
}

// validVolumeLayout checks whether the volume directory layout
// is valid to work with Docker post 1.7 or not.
func validVolumeLayout(files []os.FileInfo) bool {
	if len(files) == 1 && files[0].Name() == local.VolumeDataPathName && files[0].IsDir() {
		return true
	}

	if len(files) != 2 {
		return false
	}

	for _, f := range files {
		if f.Name() == "config.json" ||
			(f.Name() == local.VolumeDataPathName && f.Mode()&os.ModeSymlink == os.ModeSymlink) {
			// Old volume configuration, we ignore it
			continue
		}
		return false
	}

	return true
}

// verifyVolumesInfo ports volumes configured for the containers pre docker 1.7.
// It reads the container configuration and creates valid mount points for the old volumes.
func (daemon *Daemon) verifyVolumesInfo(container *Container) error {
	// Inspect old structures only when we're upgrading from old versions
	// to versions >= 1.7 and the MountPoints has not been populated with volumes data.
	if len(container.MountPoints) == 0 && len(container.Volumes) > 0 {
		for destination, hostPath := range container.Volumes {
			vfsPath := filepath.Join(daemon.root, "vfs", "dir")
			rw := container.VolumesRW != nil && container.VolumesRW[destination]

			if strings.HasPrefix(hostPath, vfsPath) {
				id := filepath.Base(hostPath)
				if err := migrateVolume(id, hostPath); err != nil {
					return err
				}
				container.addLocalMountPoint(id, destination, rw)
			} else { // Bind mount
				id, source, err := parseVolumeSource(hostPath)
				// We should not find an error here coming
				// from the old configuration, but who knows.
				if err != nil {
					return err
				}
				container.addBindMountPoint(id, source, destination, rw)
			}
		}
	} else if len(container.MountPoints) > 0 {
		// Volumes created with a Docker version >= 1.7. We verify integrity in case of data created
		// with Docker 1.7 RC versions that put the information in
		// DOCKER_ROOT/volumes/VOLUME_ID rather than DOCKER_ROOT/volumes/VOLUME_ID/_container_data.
		l, err := getVolumeDriver(volume.DefaultDriverName)
		if err != nil {
			return err
		}

		for _, m := range container.MountPoints {
			if m.Driver != volume.DefaultDriverName {
				continue
			}
			dataPath := l.(*local.Root).DataPath(m.Name)
			volumePath := filepath.Dir(dataPath)

			d, err := ioutil.ReadDir(volumePath)
			if err != nil {
				// If the volume directory doesn't exist yet it will be recreated,
				// so we only return the error when there is a different issue.
				if !os.IsNotExist(err) {
					return err
				}
				// Do not check when the volume directory does not exist.
				continue
			}
			if validVolumeLayout(d) {
				continue
			}

			if err := os.Mkdir(dataPath, 0755); err != nil {
				return err
			}

			// Move data inside the data directory
			for _, f := range d {
				oldp := filepath.Join(volumePath, f.Name())
				newp := filepath.Join(dataPath, f.Name())
				if err := os.Rename(oldp, newp); err != nil {
					logrus.Errorf("Unable to move %s to %s\n", oldp, newp)
				}
			}
		}

		return container.ToDisk()
	}

	return nil
}

func parseVolumesFrom(spec string) (string, string, error) {
	if len(spec) == 0 {
		return "", "", fmt.Errorf("malformed volumes-from specification: %s", spec)
	}

	specParts := strings.SplitN(spec, ":", 2)
	id := specParts[0]
	mode := "rw"

	if len(specParts) == 2 {
		mode = specParts[1]
		if isValid, _ := volume.ValidateMountMode(mode); !isValid {
			return "", "", fmt.Errorf("invalid mode for volumes-from: %s", mode)
		}
	}
	return id, mode, nil
}

// registerMountPoints initializes the container mount points with the configured volumes and bind mounts.
// It follows the next sequence to decide what to mount in each final destination:
//
// 1. Select the previously configured mount points for the containers, if any.
// 2. Select the volumes mounted from another containers. Overrides previously configured mount point destination.
// 3. Select the bind mounts set by the client. Overrides previously configured mount point destinations.
func (daemon *Daemon) registerMountPoints(container *Container, hostConfig *runconfig.HostConfig) error {
	binds := map[string]bool{}
	mountPoints := map[string]*mountPoint{}

	// 1. Read already configured mount points.
	for name, point := range container.MountPoints {
		mountPoints[name] = point
	}

	// 2. Read volumes from other containers.
	for _, v := range hostConfig.VolumesFrom {
		containerID, mode, err := parseVolumesFrom(v)
		if err != nil {
			return err
		}

		c, err := daemon.Get(containerID)
		if err != nil {
			return err
		}

		for _, m := range c.MountPoints {
			cp := &mountPoint{
				Name:        m.Name,
				Source:      m.Source,
				RW:          m.RW && volume.ReadWrite(mode),
				Driver:      m.Driver,
				Destination: m.Destination,
			}

			if len(cp.Source) == 0 {
				v, err := createVolume(cp.Name, cp.Driver)
				if err != nil {
					return err
				}
				cp.Volume = v
			}

			mountPoints[cp.Destination] = cp
		}
	}

	// 3. Read bind mounts
	for _, b := range hostConfig.Binds {
		// #10618
		bind, err := parseBindMount(b, container.MountLabel, container.Config)
		if err != nil {
			return err
		}

		if binds[bind.Destination] {
			return fmt.Errorf("Duplicate bind mount %s", bind.Destination)
		}

		if len(bind.Name) > 0 && len(bind.Driver) > 0 {
			// create the volume
			v, err := createVolume(bind.Name, bind.Driver)
			if err != nil {
				return err
			}
			bind.Volume = v
			bind.Source = v.Path()
			// Since this is just a named volume and not a typical bind, set to shared mode `z`
			if bind.Mode == "" {
				bind.Mode = "z"
			}
		}

		if err := label.Relabel(bind.Source, container.MountLabel, bind.Mode); err != nil {
			return err
		}
		binds[bind.Destination] = true
		mountPoints[bind.Destination] = bind
	}

	// Keep backwards compatible structures
	bcVolumes := map[string]string{}
	bcVolumesRW := map[string]bool{}
	for _, m := range mountPoints {
		if m.BackwardsCompatible() {
			bcVolumes[m.Destination] = m.Path()
			bcVolumesRW[m.Destination] = m.RW
		}
	}

	container.Lock()
	container.MountPoints = mountPoints
	container.Volumes = bcVolumes
	container.VolumesRW = bcVolumesRW
	container.Unlock()

	return nil
}

func createVolume(name, driverName string) (volume.Volume, error) {
	vd, err := getVolumeDriver(driverName)
	if err != nil {
		return nil, err
	}
	return vd.Create(name)
}

func removeVolume(v volume.Volume) error {
	vd, err := getVolumeDriver(v.DriverName())
	if err != nil {
		return nil
	}
	return vd.Remove(v)
}

func getVolumeDriver(name string) (volume.Driver, error) {
	if name == "" {
		name = volume.DefaultDriverName
	}
	return volumedrivers.Lookup(name)
}

func parseVolumeSource(spec string) (string, string, error) {
	if !filepath.IsAbs(spec) {
		return spec, "", nil
	}

	return "", spec, nil
}

// BackwardsCompatible decides whether this mount point can be
// used in old versions of Docker or not.
// Only bind mounts and local volumes can be used in old versions of Docker.
func (m *mountPoint) BackwardsCompatible() bool {
	return len(m.Source) > 0 || m.Driver == volume.DefaultDriverName
}
