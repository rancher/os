// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/coreos-cloudinit/config"
)

func NewUnitManager(root string) UnitManager {
	return &systemd{root}
}

type systemd struct {
	root string
}

// fakeMachineID is placed on non-usr CoreOS images and should
// never be used as a true MachineID
const fakeMachineID = "42000000000000000000000000000042"

// PlaceUnit writes a unit file at its desired destination, creating parent
// directories as necessary.
func (s *systemd) PlaceUnit(u Unit) error {
	file := File{config.File{
		Path:               u.Destination(s.root),
		Content:            u.Content,
		RawFilePermissions: "0644",
	}}

	_, err := WriteFile(&file, "/")
	return err
}

// PlaceUnitDropIn writes a unit drop-in file at its desired destination,
// creating parent directories as necessary.
func (s *systemd) PlaceUnitDropIn(u Unit, d config.UnitDropIn) error {
	file := File{config.File{
		Path:               u.DropInDestination(s.root, d),
		Content:            d.Content,
		RawFilePermissions: "0644",
	}}

	_, err := WriteFile(&file, "/")
	return err
}

func (s *systemd) EnableUnitFile(u Unit) error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}

	units := []string{u.Name}
	_, _, err = conn.EnableUnitFiles(units, u.Runtime, true)
	return err
}

func (s *systemd) RunUnitCommand(u Unit, c string) (string, error) {
	conn, err := dbus.New()
	if err != nil {
		return "", err
	}

	var fn func(string, string) (string, error)
	switch c {
	case "start":
		fn = conn.StartUnit
	case "stop":
		fn = conn.StopUnit
	case "restart":
		fn = conn.RestartUnit
	case "reload":
		fn = conn.ReloadUnit
	case "try-restart":
		fn = conn.TryRestartUnit
	case "reload-or-restart":
		fn = conn.ReloadOrRestartUnit
	case "reload-or-try-restart":
		fn = conn.ReloadOrTryRestartUnit
	default:
		return "", fmt.Errorf("Unsupported systemd command %q", c)
	}

	return fn(u.Name, "replace")
}

func (s *systemd) DaemonReload() error {
	conn, err := dbus.New()
	if err != nil {
		return err
	}

	return conn.Reload()
}

// MaskUnit masks the given Unit by symlinking its unit file to
// /dev/null, analogous to `systemctl mask`.
// N.B.: Unlike `systemctl mask`, this function will *remove any existing unit
// file at the location*, to ensure that the mask will succeed.
func (s *systemd) MaskUnit(u Unit) error {
	masked := u.Destination(s.root)
	if _, err := os.Stat(masked); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(masked), os.FileMode(0755)); err != nil {
			return err
		}
	} else if err := os.Remove(masked); err != nil {
		return err
	}
	return os.Symlink("/dev/null", masked)
}

// UnmaskUnit is analogous to systemd's unit_file_unmask. If the file
// associated with the given Unit is empty or appears to be a symlink to
// /dev/null, it is removed.
func (s *systemd) UnmaskUnit(u Unit) error {
	masked := u.Destination(s.root)
	ne, err := nullOrEmpty(masked)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !ne {
		log.Printf("%s is not null or empty, refusing to unmask", masked)
		return nil
	}
	return os.Remove(masked)
}

// nullOrEmpty checks whether a given path appears to be an empty regular file
// or a symlink to /dev/null
func nullOrEmpty(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	m := fi.Mode()
	if m.IsRegular() && fi.Size() <= 0 {
		return true, nil
	}
	if m&os.ModeCharDevice > 0 {
		return true, nil
	}
	return false, nil
}

func ExecuteScript(scriptPath string) (string, error) {
	props := []dbus.Property{
		dbus.PropDescription("Unit generated and executed by coreos-cloudinit on behalf of user"),
		dbus.PropExecStart([]string{"/bin/bash", scriptPath}, false),
	}

	base := path.Base(scriptPath)
	name := fmt.Sprintf("coreos-cloudinit-%s.service", base)

	log.Printf("Creating transient systemd unit '%s'", name)

	conn, err := dbus.New()
	if err != nil {
		return "", err
	}

	_, err = conn.StartTransientUnit(name, "replace", props...)
	return name, err
}

func SetHostname(hostname string) error {
	return exec.Command("hostnamectl", "set-hostname", hostname).Run()
}

func Hostname() (string, error) {
	return os.Hostname()
}

func MachineID(root string) string {
	contents, _ := ioutil.ReadFile(path.Join(root, "etc", "machine-id"))
	id := strings.TrimSpace(string(contents))

	if id == fakeMachineID {
		id = ""
	}

	return id
}
