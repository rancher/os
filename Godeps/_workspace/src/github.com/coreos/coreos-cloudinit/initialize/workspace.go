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
	"io/ioutil"
	"path"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/system"
)

func PrepWorkspace(workspace string) error {
	if err := system.EnsureDirectoryExists(workspace); err != nil {
		return err
	}

	scripts := path.Join(workspace, "scripts")
	if err := system.EnsureDirectoryExists(scripts); err != nil {
		return err
	}

	return nil
}

func PersistScriptInWorkspace(script config.Script, workspace string) (string, error) {
	scriptsPath := path.Join(workspace, "scripts")
	tmp, err := ioutil.TempFile(scriptsPath, "")
	if err != nil {
		return "", err
	}
	tmp.Close()

	relpath := strings.TrimPrefix(tmp.Name(), workspace)

	file := system.File{File: config.File{
		Path:               relpath,
		RawFilePermissions: "0744",
		Content:            string(script),
	}}

	return system.WriteFile(&file, workspace)
}

func PersistUnitNameInWorkspace(name string, workspace string) error {
	file := system.File{File: config.File{
		Path:               path.Join("scripts", "unit-name"),
		RawFilePermissions: "0644",
		Content:            name,
	}}
	_, err := system.WriteFile(&file, workspace)
	return err
}
