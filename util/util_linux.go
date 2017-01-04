// +build linux

package util

import (
	"os"
	"syscall"
	"strings"

	"github.com/docker/docker/pkg/mount"
)

func mountProc() error {
	if _, err := os.Stat("/proc/self/mountinfo"); os.IsNotExist(err) {
		if _, err := os.Stat("/proc"); os.IsNotExist(err) {
			if err = os.Mkdir("/proc", 0755); err != nil {
				return err
			}
		}

		if err := syscall.Mount("none", "/proc", "proc", 0, ""); err != nil {
			return err
		}
	}

	return nil
}

func Mount(device, directory, fsType string, options ...string) error {
	if err := mountProc(); err != nil {
		return nil
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	return mount.Mount(device, directory, fsType, strings.Join(options, ","))
}
