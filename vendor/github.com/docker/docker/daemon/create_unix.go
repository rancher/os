// +build !windows

package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/stringid"
	"github.com/docker/docker/runconfig"
	"github.com/opencontainers/runc/libcontainer/label"
)

// createContainerPlatformSpecificSettings performs platform specific container create functionality
func createContainerPlatformSpecificSettings(container *Container, config *runconfig.Config) error {
	for spec := range config.Volumes {
		var (
			name, destination string
			parts             = strings.Split(spec, ":")
		)
		switch len(parts) {
		case 2:
			name, destination = parts[0], filepath.Clean(parts[1])
		default:
			name = stringid.GenerateNonCryptoID()
			destination = filepath.Clean(parts[0])
		}
		// Skip volumes for which we already have something mounted on that
		// destination because of a --volume-from.
		if container.isDestinationMounted(destination) {
			continue
		}
		path, err := container.GetResourcePath(destination)
		if err != nil {
			return err
		}

		stat, err := os.Stat(path)
		if err == nil && !stat.IsDir() {
			return fmt.Errorf("cannot mount volume over existing file, file exists %s", path)
		}

		v, err := createVolume(name, config.VolumeDriver)
		if err != nil {
			return err
		}
		if err := label.Relabel(v.Path(), container.MountLabel, "z"); err != nil {
			return err
		}

		if err := container.copyImagePathContent(v, destination); err != nil {
			return err
		}

		container.addMountPointWithVolume(destination, v, true)
	}
	return nil
}
