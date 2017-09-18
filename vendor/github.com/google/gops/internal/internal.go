package internal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

func ConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "gops"), nil
	}
	return filepath.Join("/run", "gops"), nil
}

func guessUnixHomeDir() string {
	usr, err := user.Current()
	if err == nil {
		return usr.HomeDir
	}
	return os.Getenv("HOME")
}

func PIDFile(pid int) (string, error) {
	gopsdir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%d", gopsdir, pid), nil
}

func GetPort(pid int) (string, error) {
	portfile, err := PIDFile(pid)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(portfile)
	if err != nil {
		return "", err
	}
	port := strings.TrimSpace(string(b))
	return port, nil
}
