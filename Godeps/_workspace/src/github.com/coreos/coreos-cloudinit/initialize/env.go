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
	"net"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/system"
)

const DefaultSSHKeyName = "coreos-cloudinit"

type Environment struct {
	root          string
	configRoot    string
	workspace     string
	sshKeyName    string
	substitutions map[string]string
}

// TODO(jonboulle): this is getting unwieldy, should be able to simplify the interface somehow
func NewEnvironment(root, configRoot, workspace, sshKeyName string, metadata datasource.Metadata) *Environment {
	firstNonNull := func(ip net.IP, env string) string {
		if ip == nil {
			return env
		}
		return ip.String()
	}
	substitutions := map[string]string{
		"$public_ipv4":  firstNonNull(metadata.PublicIPv4, os.Getenv("COREOS_PUBLIC_IPV4")),
		"$private_ipv4": firstNonNull(metadata.PrivateIPv4, os.Getenv("COREOS_PRIVATE_IPV4")),
		"$public_ipv6":  firstNonNull(metadata.PublicIPv6, os.Getenv("COREOS_PUBLIC_IPV6")),
		"$private_ipv6": firstNonNull(metadata.PrivateIPv6, os.Getenv("COREOS_PRIVATE_IPV6")),
	}
	return &Environment{root, configRoot, workspace, sshKeyName, substitutions}
}

func (e *Environment) Workspace() string {
	return path.Join(e.root, e.workspace)
}

func (e *Environment) Root() string {
	return e.root
}

func (e *Environment) ConfigRoot() string {
	return e.configRoot
}

func (e *Environment) SSHKeyName() string {
	return e.sshKeyName
}

func (e *Environment) SetSSHKeyName(name string) {
	e.sshKeyName = name
}

// Apply goes through the map of substitutions and replaces all instances of
// the keys with their respective values. It supports escaping substitutions
// with a leading '\'.
func (e *Environment) Apply(data string) string {
	for key, val := range e.substitutions {
		matchKey := strings.Replace(key, `$`, `\$`, -1)
		replKey := strings.Replace(key, `$`, `$$`, -1)

		// "key" -> "val"
		data = regexp.MustCompile(`([^\\]|^)`+matchKey).ReplaceAllString(data, `${1}`+val)
		// "\key" -> "key"
		data = regexp.MustCompile(`\\`+matchKey).ReplaceAllString(data, replKey)
	}
	return data
}

func (e *Environment) DefaultEnvironmentFile() *system.EnvFile {
	ef := system.EnvFile{
		File: &system.File{File: config.File{
			Path: "/etc/environment",
		}},
		Vars: map[string]string{},
	}
	if ip, ok := e.substitutions["$public_ipv4"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PUBLIC_IPV4"] = ip
	}
	if ip, ok := e.substitutions["$private_ipv4"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PRIVATE_IPV4"] = ip
	}
	if ip, ok := e.substitutions["$public_ipv6"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PUBLIC_IPV6"] = ip
	}
	if ip, ok := e.substitutions["$private_ipv6"]; ok && len(ip) > 0 {
		ef.Vars["COREOS_PRIVATE_IPV6"] = ip
	}
	if len(ef.Vars) == 0 {
		return nil
	} else {
		return &ef
	}
}
