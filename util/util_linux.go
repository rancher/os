// +build linux

package util

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"strings"

	"github.com/docker/docker/pkg/mount"
	"github.com/rancher/os/log"
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

func Unmount(target string) error {
	return mount.Unmount(target)
}

func Blkid(label string) (deviceName, deviceType string) {
	// Not all blkid's have `blkid -L label (see busybox/alpine)
	cmd := exec.Command("blkid")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		log.Errorf("Failed to run blkid: %s", err)
		return
	}
	r := bytes.NewReader(out)
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		//log.Debugf("blkid: %s", cmd, line)
		if !strings.Contains(line, `LABEL="`+label+`"`) {
			continue
		}
		d := strings.Split(line, ":")
		deviceName = d[0]

		s1 := strings.Split(line, `TYPE="`)
		s2 := strings.Split(s1[1], `"`)
		deviceType = s2[0]
		log.Debugf("blkid type of %s: %s", deviceName, deviceType)
		return
	}
	return
}
