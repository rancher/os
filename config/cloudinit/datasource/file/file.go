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

package file

import (
	"io/ioutil"
	"os"

	"github.com/rancher/os/config/cloudinit/datasource"
)

type LocalFile struct {
	path string
}

func NewDatasource(path string) *LocalFile {
	return &LocalFile{path}
}

func (f *LocalFile) IsAvailable() bool {
	_, err := os.Stat(f.path)
	return !os.IsNotExist(err)
}

func (f *LocalFile) AvailabilityChanges() bool {
	return true
}

func (f *LocalFile) ConfigRoot() string {
	return ""
}

func (f *LocalFile) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (f *LocalFile) FetchUserdata() ([]byte, error) {
	return ioutil.ReadFile(f.path)
}

func (f *LocalFile) Type() string {
	return "local-file"
}
