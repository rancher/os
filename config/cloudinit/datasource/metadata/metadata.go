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

package metadata

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/rancher/os/config/cloudinit/pkg"
	"github.com/rancher/os/log"
)

type Service struct {
	Root         string
	Client       pkg.Getter
	APIVersion   string
	UserdataPath string
	MetadataPath string
	lastError    error
}

func NewDatasource(root, apiVersion, userdataPath, metadataPath string, header http.Header) Service {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	return Service{root, pkg.NewHTTPClientHeader(header), apiVersion, userdataPath, metadataPath, nil}
}

func (ms Service) IsAvailable() bool {
	_, ms.lastError = ms.Client.Get(ms.Root + ms.APIVersion)
	if ms.lastError != nil {
		log.Errorf("%s: %s (lastError: %s)", "IsAvailable", ms.Root+":"+ms.UserdataPath, ms.lastError)
	}
	return (ms.lastError == nil)
}

func (ms *Service) Finish() error {
	return nil
}

func (ms *Service) String() string {
	return fmt.Sprintf("%s: %s (lastError: %s)", "metadata", ms.Root+ms.UserdataPath, ms.lastError)
}

func (ms Service) AvailabilityChanges() bool {
	return true
}

func (ms Service) ConfigRoot() string {
	return ms.Root
}

func (ms Service) FetchUserdata() ([]byte, error) {
	return ms.FetchData(ms.UserdataURL())
}

func (ms Service) FetchData(url string) ([]byte, error) {
	if data, err := ms.Client.GetRetry(url); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrNotFound); ok {
		return []byte{}, nil
	} else {
		return data, err
	}
}

func (ms Service) MetadataURL() string {
	return (ms.Root + ms.MetadataPath)
}

func (ms Service) UserdataURL() string {
	return (ms.Root + ms.UserdataPath)
}

func (ms Service) FetchAttributes(key string) ([]string, error) {
	url := ms.MetadataURL() + key
	resp, err := ms.FetchData(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

func (ms Service) FetchAttribute(key string) (string, error) {
	attrs, err := ms.FetchAttributes(key)
	if err == nil && len(attrs) > 0 {
		return attrs[0], nil
	}
	return "", err
}
