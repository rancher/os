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
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

const (
	base          = "# a file\nFOO=base\n\nBAR= hi there\n"
	baseNoNewline = "# a file\nFOO=base\n\nBAR= hi there"
	baseDos       = "# a file\r\nFOO=base\r\n\r\nBAR= hi there\r\n"
	expectUpdate  = "# a file\nFOO=test\n\nBAR= hi there\nNEW=a value\n"
	expectCreate  = "FOO=test\nNEW=a value\n"
)

var (
	valueUpdate = map[string]string{
		"FOO": "test",
		"NEW": "a value",
	}
	valueNoop = map[string]string{
		"FOO": "base",
	}
	valueEmpty   = map[string]string{}
	valueInvalid = map[string]string{
		"FOO-X": "test",
	}
)

func TestWriteEnvFileUpdate(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(base), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueUpdate,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != expectUpdate {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino == newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was not replaced: %s", fullPath)
	}
}

func TestWriteEnvFileUpdateNoNewline(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(baseNoNewline), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueUpdate,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != expectUpdate {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino == newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was not replaced: %s", fullPath)
	}
}

func TestWriteEnvFileCreate(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueUpdate,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != expectCreate {
		t.Fatalf("File has incorrect contents: %q", contents)
	}
}

func TestWriteEnvFileNoop(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(base), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueNoop,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != base {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino != newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was replaced: %s", fullPath)
	}
}

func TestWriteEnvFileUpdateDos(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(baseDos), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueUpdate,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != expectUpdate {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino == newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was not replaced: %s", fullPath)
	}
}

// A middle ground noop, values are unchanged but we did have a value.
// Seems reasonable to rewrite the file in Unix format anyway.
func TestWriteEnvFileDos2Unix(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(baseDos), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueNoop,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != base {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino == newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was not replaced: %s", fullPath)
	}
}

// If it really is a noop (structure is empty) don't even do dos2unix
func TestWriteEnvFileEmpty(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(baseDos), 0644)

	oldStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueEmpty,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	if string(contents) != baseDos {
		t.Fatalf("File has incorrect contents: %q", contents)
	}

	newStat, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if oldStat.Sys().(*syscall.Stat_t).Ino != newStat.Sys().(*syscall.Stat_t).Ino {
		t.Fatalf("File was replaced: %s", fullPath)
	}
}

// no point in creating empty files
func TestWriteEnvFileEmptyNoCreate(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueEmpty,
	}

	err = WriteEnvFile(&ef, dir)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err == nil {
		t.Fatalf("File has incorrect contents: %q", contents)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Unexpected error while reading file: %v", err)
	}
}

func TestWriteEnvFilePermFailure(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"
	fullPath := path.Join(dir, name)
	ioutil.WriteFile(fullPath, []byte(base), 0000)

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueUpdate,
	}

	err = WriteEnvFile(&ef, dir)
	if !os.IsPermission(err) {
		t.Fatalf("Not a pemission denied error: %v", err)
	}
}

func TestWriteEnvFileNameFailure(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	name := "foo.conf"

	ef := EnvFile{
		File: &File{config.File{
			Path: name,
		}},
		Vars: valueInvalid,
	}

	err = WriteEnvFile(&ef, dir)
	if err == nil || !strings.HasPrefix(err.Error(), "Invalid name") {
		t.Fatalf("Not an invalid name error: %v", err)
	}
}
