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
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

const (
	locksmithUnit    = "locksmithd.service"
	updateEngineUnit = "update-engine.service"
)

// Update is a top-level structure which contains its underlying configuration,
// config.Update, a function for reading the configuration (the default
// implementation reading from the filesystem), and provides the system-specific
// File() and Unit().
type Update struct {
	ReadConfig func() (io.Reader, error)
	config.Update
}

func DefaultReadConfig() (io.Reader, error) {
	etcUpdate := path.Join("/etc", "coreos", "update.conf")
	usrUpdate := path.Join("/usr", "share", "coreos", "update.conf")

	f, err := os.Open(etcUpdate)
	if os.IsNotExist(err) {
		f, err = os.Open(usrUpdate)
	}
	return f, err
}

// File generates an `/etc/coreos/update.conf` file (if any update
// configuration options are set in cloud-config) by either rewriting the
// existing file on disk, or starting from `/usr/share/coreos/update.conf`
func (uc Update) File() (*File, error) {
	if config.IsZero(uc.Update) {
		return nil, nil
	}
	if err := config.AssertStructValid(uc.Update); err != nil {
		return nil, err
	}

	// Generate the list of possible substitutions to be performed based on the options that are configured
	subs := map[string]string{}
	uct := reflect.TypeOf(uc.Update)
	ucv := reflect.ValueOf(uc.Update)
	for i := 0; i < uct.NumField(); i++ {
		val := ucv.Field(i).String()
		if val == "" {
			continue
		}
		env := uct.Field(i).Tag.Get("env")
		subs[env] = fmt.Sprintf("%s=%s", env, val)
	}

	conf, err := uc.ReadConfig()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(conf)

	var out string
	for scanner.Scan() {
		line := scanner.Text()
		for env, value := range subs {
			if strings.HasPrefix(line, env) {
				line = value
				delete(subs, env)
				break
			}
		}
		out += line
		out += "\n"
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	for _, key := range sortedKeys(subs) {
		out += subs[key]
		out += "\n"
	}

	return &File{config.File{
		Path:               path.Join("etc", "coreos", "update.conf"),
		RawFilePermissions: "0644",
		Content:            out,
	}}, nil
}

// Units generates units for the cloud-init initializer to act on:
// - a locksmith Unit, if "reboot-strategy" was set in cloud-config
// - an update_engine Unit, if "group" or "server" was set in cloud-config
func (uc Update) Units() []Unit {
	var units []Unit
	if uc.Update.RebootStrategy != "" {
		ls := &Unit{config.Unit{
			Name:    locksmithUnit,
			Command: "restart",
			Mask:    false,
			Runtime: true,
		}}

		if uc.Update.RebootStrategy == "off" {
			ls.Command = "stop"
			ls.Mask = true
		}
		units = append(units, *ls)
	}

	if uc.Update.Group != "" || uc.Update.Server != "" {
		ue := Unit{config.Unit{
			Name:    updateEngineUnit,
			Command: "restart",
		}}
		units = append(units, ue)
	}

	return units
}

func sortedKeys(m map[string]string) (keys []string) {
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return
}
