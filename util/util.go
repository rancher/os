package util

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

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

func Mount(device, directory, fsType, options string) error {
	if err := mountProc(); err != nil {
		return nil
	}

	if _, err := os.Stat(directory); os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	}

	return mount.Mount(device, directory, fsType, options)
}

func Remount(directory, options string) error {
	return mount.Mount("", directory, "", fmt.Sprintf("remount,%s", options))
}

func ExtractTar(archive string, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	input := tar.NewReader(f)

	for {
		header, err := input.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if header == nil {
			break
		}

		fileInfo := header.FileInfo()
		fileName := path.Join(dest, header.Name)
		if fileInfo.IsDir() {
			//log.Debugf("DIR : %s", fileName)
			err = os.MkdirAll(fileName, fileInfo.Mode())
			if err != nil {
				return err
			}
		} else {
			//log.Debugf("FILE: %s", fileName)
			destFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
			if err != nil {
				return err
			}

			_, err = io.Copy(destFile, input)
			// Not deferring, concerned about holding open too many files
			destFile.Close()

			if err != nil {
				return err
			}
		}
	}

	return nil
}
