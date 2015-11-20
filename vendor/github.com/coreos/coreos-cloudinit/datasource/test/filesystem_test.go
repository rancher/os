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
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestReadFile(t *testing.T) {
	tests := []struct {
		filesystem MockFilesystem

		filename string
		contents string
		err      error
	}{
		{
			filename: "dne",
			err:      os.ErrNotExist,
		},
		{
			filesystem: MockFilesystem{
				"exists": File{Contents: "hi"},
			},
			filename: "exists",
			contents: "hi",
		},
		{
			filesystem: MockFilesystem{
				"dir": File{Directory: true},
			},
			filename: "dir",
			err:      errors.New("read dir: is a directory"),
		},
	}

	for i, tt := range tests {
		contents, err := tt.filesystem.ReadFile(tt.filename)
		if tt.contents != string(contents) {
			t.Errorf("bad contents (test %d): want %q, got %q", i, tt.contents, string(contents))
		}
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (test %d): want %v, got %v", i, tt.err, err)
		}
	}
}

func TestNewMockFilesystem(t *testing.T) {
	tests := []struct {
		files []File

		filesystem MockFilesystem
	}{
		{
			filesystem: MockFilesystem{},
		},
		{
			files: []File{File{Path: "file"}},
			filesystem: MockFilesystem{
				"file": File{Path: "file"},
			},
		},
		{
			files: []File{File{Path: "/file"}},
			filesystem: MockFilesystem{
				"/file": File{Path: "/file"},
			},
		},
		{
			files: []File{File{Path: "/dir/file"}},
			filesystem: MockFilesystem{
				"/dir":      File{Path: "/dir", Directory: true},
				"/dir/file": File{Path: "/dir/file"},
			},
		},
		{
			files: []File{File{Path: "/dir/dir/file"}},
			filesystem: MockFilesystem{
				"/dir":          File{Path: "/dir", Directory: true},
				"/dir/dir":      File{Path: "/dir/dir", Directory: true},
				"/dir/dir/file": File{Path: "/dir/dir/file"},
			},
		},
		{
			files: []File{File{Path: "/dir/dir/dir", Directory: true}},
			filesystem: MockFilesystem{
				"/dir":         File{Path: "/dir", Directory: true},
				"/dir/dir":     File{Path: "/dir/dir", Directory: true},
				"/dir/dir/dir": File{Path: "/dir/dir/dir", Directory: true},
			},
		},
	}

	for i, tt := range tests {
		filesystem := NewMockFilesystem(tt.files...)
		if !reflect.DeepEqual(tt.filesystem, filesystem) {
			t.Errorf("bad filesystem (test %d): want %#v, got %#v", i, tt.filesystem, filesystem)
		}
	}
}
