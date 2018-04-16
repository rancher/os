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

package test

import (
	"fmt"
	"os"
	"path"
)

type MockFilesystem map[string]File

type File struct {
	Path      string
	Contents  string
	Directory bool
}

func (m MockFilesystem) ReadFile(filename string) ([]byte, error) {
	if f, ok := m[path.Clean(filename)]; ok {
		if f.Directory {
			return nil, fmt.Errorf("read %s: is a directory", filename)
		}
		return []byte(f.Contents), nil
	}
	return nil, os.ErrNotExist
}

func NewMockFilesystem(files ...File) MockFilesystem {
	fs := MockFilesystem{}
	for _, file := range files {
		fs[file.Path] = file

		// Create the directories leading up to the file
		p := path.Dir(file.Path)
		for p != "/" && p != "." {
			if f, ok := fs[p]; ok && !f.Directory {
				panic(fmt.Sprintf("%q already exists and is not a directory (%#v)", p, f))
			}
			fs[p] = File{Path: p, Directory: true}
			p = path.Dir(p)
		}
	}
	return fs
}
