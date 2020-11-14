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
	"fmt"

	"github.com/burmilla/os/config/cloudinit/datasource"
	"github.com/burmilla/os/config/cloudinit/pkg"
	"github.com/burmilla/os/pkg/util/network"
)

type RemoteFile struct {
	url       string
	lastError error
}

func NewDatasource(url string) *RemoteFile {
	return &RemoteFile{url, nil}
}

func (f *RemoteFile) IsAvailable() bool {
	network.SetProxyEnvironmentVariables()
	client := pkg.NewHTTPClient()
	_, f.lastError = client.GetRetry(f.url)
	return (f.lastError == nil)
}

func (f *RemoteFile) Finish() error {
	return nil
}

func (f *RemoteFile) String() string {
	return fmt.Sprintf("%s: %s (lastError: %v)", f.Type(), f.url, f.lastError)
}

func (f *RemoteFile) AvailabilityChanges() bool {
	return false
	// TODO: we should trigger something to change the network state
	/*	if f.lastError != nil {
			// if we have a Network error, then we should retry.
			// otherwise, we've made a request to the server, and its said nope.
			if _, ok := f.lastError.(pkg.ErrNetwork); !ok {
				return false
			}
		}
		return true
	*/
}

func (f *RemoteFile) ConfigRoot() string {
	return ""
}

func (f *RemoteFile) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (f *RemoteFile) FetchUserdata() ([]byte, error) {
	client := pkg.NewHTTPClient()
	return client.GetRetry(f.url)
}

func (f *RemoteFile) Type() string {
	return "url"
}
