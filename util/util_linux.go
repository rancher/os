// +build linux

package util

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/SvenDowideit/cpuid"
	"github.com/docker/docker/pkg/mount"
)

const (
	tmpfsMagic int64 = 0x01021994
	ramfsMagic int64 = 0x858458f6
)

func RootFsIsNotReal() bool {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	return int64(stat.Type) == tmpfsMagic || int64(stat.Type) == ramfsMagic
}

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

func Mount(device, target, fsType, options string) error {
	if err := mountProc(); err != nil {
		return nil
	}

	bindMount := false
	for _, v := range strings.Split(options, ",") {
		if v == "bind" {
			bindMount = true
			break
		}
	}

	if bindMount {
		deviceInfo, err := os.Stat(device)
		if err != nil {
			return err
		}
		mode := deviceInfo.Mode()

		switch {
		case mode.IsDir():
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case mode.IsRegular():
			err := os.MkdirAll(filepath.Dir(target), 0755)
			if err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE, mode&os.ModePerm)
			if err != nil {
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		default:
			return os.ErrInvalid
		}
	} else {
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return err
		}

		if fsType == "auto" || fsType == "" {
			inferredType, err := GetFsType(device)
			if err != nil {
				return err
			}
			fsType = inferredType
		}
	}

	return mount.Mount(device, target, fsType, options)
}

func Unmount(target string) error {
	return mount.Unmount(target)
}

func Blkid(label string) (deviceName, deviceType string, err error) {
	// Not all blkid's have `blkid -L label (see busybox/alpine)
	cmd := exec.Command("blkid")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return
	}
	r := bytes.NewReader(out)
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if !strings.Contains(line, `LABEL="`+label+`"`) {
			continue
		}
		d := strings.Split(line, ":")
		deviceName = d[0]

		s1 := strings.Split(line, `TYPE="`)
		s2 := strings.Split(s1[1], `"`)
		deviceType = s2[0]
		return
	}
	return
}

// GetHypervisor tries to detect if we're running in a VM, and returns a string for its type
func GetHypervisor() string {
	return cpuid.CPU.HypervisorName
}
