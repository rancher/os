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

package initialize

import (
	"errors"
	"fmt"
	"log"
	"path"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/system"
)

// CloudConfigFile represents a CoreOS specific configuration option that can generate
// an associated system.File to be written to disk
type CloudConfigFile interface {
	// File should either return (*system.File, error), or (nil, nil) if nothing
	// needs to be done for this configuration option.
	File() (*system.File, error)
}

// CloudConfigUnit represents a CoreOS specific configuration option that can generate
// associated system.Units to be created/enabled appropriately
type CloudConfigUnit interface {
	Units() []system.Unit
}

// Apply renders a CloudConfig to an Environment. This can involve things like
// configuring the hostname, adding new users, writing various configuration
// files to disk, and manipulating systemd services.
func Apply(cfg config.CloudConfig, ifaces []network.InterfaceGenerator, env *Environment) error {
	if cfg.Hostname != "" {
		if err := system.SetHostname(cfg.Hostname); err != nil {
			return err
		}
		log.Printf("Set hostname to %s", cfg.Hostname)
	}

	for _, user := range cfg.Users {
		if user.Name == "" {
			log.Printf("User object has no 'name' field, skipping")
			continue
		}

		if system.UserExists(&user) {
			log.Printf("User '%s' exists, ignoring creation-time fields", user.Name)
			if user.PasswordHash != "" {
				log.Printf("Setting '%s' user's password", user.Name)
				if err := system.SetUserPassword(user.Name, user.PasswordHash); err != nil {
					log.Printf("Failed setting '%s' user's password: %v", user.Name, err)
					return err
				}
			}
		} else {
			log.Printf("Creating user '%s'", user.Name)
			if err := system.CreateUser(&user); err != nil {
				log.Printf("Failed creating user '%s': %v", user.Name, err)
				return err
			}
		}

		if len(user.SSHAuthorizedKeys) > 0 {
			log.Printf("Authorizing %d SSH keys for user '%s'", len(user.SSHAuthorizedKeys), user.Name)
			if err := system.AuthorizeSSHKeys(user.Name, env.SSHKeyName(), user.SSHAuthorizedKeys); err != nil {
				return err
			}
		}
		if user.SSHImportGithubUser != "" {
			log.Printf("Authorizing github user %s SSH keys for CoreOS user '%s'", user.SSHImportGithubUser, user.Name)
			if err := SSHImportGithubUser(user.Name, user.SSHImportGithubUser); err != nil {
				return err
			}
		}
		for _, u := range user.SSHImportGithubUsers {
			log.Printf("Authorizing github user %s SSH keys for CoreOS user '%s'", u, user.Name)
			if err := SSHImportGithubUser(user.Name, u); err != nil {
				return err
			}
		}
		if user.SSHImportURL != "" {
			log.Printf("Authorizing SSH keys for CoreOS user '%s' from '%s'", user.Name, user.SSHImportURL)
			if err := SSHImportKeysFromURL(user.Name, user.SSHImportURL); err != nil {
				return err
			}
		}
	}

	if len(cfg.SSHAuthorizedKeys) > 0 {
		err := system.AuthorizeSSHKeys("core", env.SSHKeyName(), cfg.SSHAuthorizedKeys)
		if err == nil {
			log.Printf("Authorized SSH keys for core user")
		} else {
			return err
		}
	}

	var writeFiles []system.File
	for _, file := range cfg.WriteFiles {
		writeFiles = append(writeFiles, system.File{File: file})
	}

	for _, ccf := range []CloudConfigFile{
		system.OEM{OEM: cfg.CoreOS.OEM},
		system.Update{Update: cfg.CoreOS.Update, ReadConfig: system.DefaultReadConfig},
		system.EtcHosts{EtcHosts: cfg.ManageEtcHosts},
		system.Flannel{Flannel: cfg.CoreOS.Flannel},
	} {
		f, err := ccf.File()
		if err != nil {
			return err
		}
		if f != nil {
			writeFiles = append(writeFiles, *f)
		}
	}

	var units []system.Unit
	for _, u := range cfg.CoreOS.Units {
		units = append(units, system.Unit{Unit: u})
	}

	for _, ccu := range []CloudConfigUnit{
		system.Etcd{Etcd: cfg.CoreOS.Etcd},
		system.Etcd2{Etcd2: cfg.CoreOS.Etcd2},
		system.Fleet{Fleet: cfg.CoreOS.Fleet},
		system.Locksmith{Locksmith: cfg.CoreOS.Locksmith},
		system.Update{Update: cfg.CoreOS.Update, ReadConfig: system.DefaultReadConfig},
	} {
		units = append(units, ccu.Units()...)
	}

	wroteEnvironment := false
	for _, file := range writeFiles {
		fullPath, err := system.WriteFile(&file, env.Root())
		if err != nil {
			return err
		}
		if path.Clean(file.Path) == "/etc/environment" {
			wroteEnvironment = true
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}

	if !wroteEnvironment {
		ef := env.DefaultEnvironmentFile()
		if ef != nil {
			err := system.WriteEnvFile(ef, env.Root())
			if err != nil {
				return err
			}
			log.Printf("Updated /etc/environment")
		}
	}

	if len(ifaces) > 0 {
		units = append(units, createNetworkingUnits(ifaces)...)
		if err := system.RestartNetwork(ifaces); err != nil {
			return err
		}
	}

	um := system.NewUnitManager(env.Root())
	return processUnits(units, env.Root(), um)
}

func createNetworkingUnits(interfaces []network.InterfaceGenerator) (units []system.Unit) {
	appendNewUnit := func(units []system.Unit, name, content string) []system.Unit {
		if content == "" {
			return units
		}
		return append(units, system.Unit{Unit: config.Unit{
			Name:    name,
			Runtime: true,
			Content: content,
		}})
	}
	for _, i := range interfaces {
		units = appendNewUnit(units, fmt.Sprintf("%s.netdev", i.Filename()), i.Netdev())
		units = appendNewUnit(units, fmt.Sprintf("%s.link", i.Filename()), i.Link())
		units = appendNewUnit(units, fmt.Sprintf("%s.network", i.Filename()), i.Network())
	}
	return units
}

// processUnits takes a set of Units and applies them to the given root using
// the given UnitManager. This can involve things like writing unit files to
// disk, masking/unmasking units, or invoking systemd
// commands against units. It returns any error encountered.
func processUnits(units []system.Unit, root string, um system.UnitManager) error {
	type action struct {
		unit    system.Unit
		command string
	}
	actions := make([]action, 0, len(units))
	reload := false
	restartNetworkd := false
	for _, unit := range units {
		if unit.Name == "" {
			log.Printf("Skipping unit without name")
			continue
		}

		if unit.Content != "" {
			log.Printf("Writing unit %q to filesystem", unit.Name)
			if err := um.PlaceUnit(unit); err != nil {
				return err
			}
			log.Printf("Wrote unit %q", unit.Name)
			reload = true
		}

		for _, dropin := range unit.DropIns {
			if dropin.Name != "" && dropin.Content != "" {
				log.Printf("Writing drop-in unit %q to filesystem", dropin.Name)
				if err := um.PlaceUnitDropIn(unit, dropin); err != nil {
					return err
				}
				log.Printf("Wrote drop-in unit %q", dropin.Name)
				reload = true
			}
		}

		if unit.Mask {
			log.Printf("Masking unit file %q", unit.Name)
			if err := um.MaskUnit(unit); err != nil {
				return err
			}
		} else if unit.Runtime {
			log.Printf("Ensuring runtime unit file %q is unmasked", unit.Name)
			if err := um.UnmaskUnit(unit); err != nil {
				return err
			}
		}

		if unit.Enable {
			if unit.Group() != "network" {
				log.Printf("Enabling unit file %q", unit.Name)
				if err := um.EnableUnitFile(unit); err != nil {
					return err
				}
				log.Printf("Enabled unit %q", unit.Name)
			} else {
				log.Printf("Skipping enable for network-like unit %q", unit.Name)
			}
		}

		if unit.Group() == "network" {
			restartNetworkd = true
		} else if unit.Command != "" {
			actions = append(actions, action{unit, unit.Command})
		}
	}

	if reload {
		if err := um.DaemonReload(); err != nil {
			return errors.New(fmt.Sprintf("failed systemd daemon-reload: %s", err))
		}
	}

	if restartNetworkd {
		log.Printf("Restarting systemd-networkd")
		networkd := system.Unit{Unit: config.Unit{Name: "systemd-networkd.service"}}
		res, err := um.RunUnitCommand(networkd, "restart")
		if err != nil {
			return err
		}
		log.Printf("Restarted systemd-networkd (%s)", res)
	}

	for _, action := range actions {
		log.Printf("Calling unit command %q on %q'", action.command, action.unit.Name)
		res, err := um.RunUnitCommand(action.unit, action.command)
		if err != nil {
			return err
		}
		log.Printf("Result of %q on %q: %s", action.command, action.unit.Name, res)
	}

	return nil
}
