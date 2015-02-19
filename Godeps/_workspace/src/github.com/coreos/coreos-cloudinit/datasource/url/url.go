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

package url

import (
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/pkg"
)

type remoteFile struct {
	url string
}

func NewDatasource(url string) *remoteFile {
	return &remoteFile{url}
}

func (f *remoteFile) IsAvailable() bool {
	client := pkg.NewHttpClient()
	_, err := client.Get(f.url)
	return (err == nil)
}

func (f *remoteFile) AvailabilityChanges() bool {
	return true
}

func (f *remoteFile) ConfigRoot() string {
	return ""
}

func (f *remoteFile) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (f *remoteFile) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	return client.GetRetry(f.url)
}

func (f *remoteFile) Type() string {
	return "url"
}
