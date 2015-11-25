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
	"bytes"
	"fmt"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource/metadata/test"
	"github.com/coreos/coreos-cloudinit/pkg"
)

func TestAvailabilityChanges(t *testing.T) {
	want := true
	if ac := (MetadataService{}).AvailabilityChanges(); ac != want {
		t.Fatalf("bad AvailabilityChanges: want %t, got %t", want, ac)
	}
}

func TestIsAvailable(t *testing.T) {
	for _, tt := range []struct {
		root       string
		apiVersion string
		resources  map[string]string
		expect     bool
	}{
		{
			root:       "/",
			apiVersion: "2009-04-04",
			resources: map[string]string{
				"/2009-04-04": "",
			},
			expect: true,
		},
		{
			root:      "/",
			resources: map[string]string{},
			expect:    false,
		},
	} {
		service := &MetadataService{
			Root:       tt.root,
			Client:     &test.HttpClient{Resources: tt.resources, Err: nil},
			ApiVersion: tt.apiVersion,
		}
		if a := service.IsAvailable(); a != tt.expect {
			t.Fatalf("bad isAvailable (%q): want %t, got %t", tt.resources, tt.expect, a)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		userdataPath string
		resources    map[string]string
		userdata     []byte
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			userdataPath: "2009-04-04/user-data",
			resources: map[string]string{
				"/2009-04-04/user-data": "hello",
			},
			userdata: []byte("hello"),
		},
		{
			root:      "/",
			clientErr: pkg.ErrNotFound{Err: fmt.Errorf("test not found error")},
			userdata:  []byte{},
		},
		{
			root:      "/",
			clientErr: pkg.ErrTimeout{Err: fmt.Errorf("test timeout error")},
			expectErr: pkg.ErrTimeout{Err: fmt.Errorf("test timeout error")},
		},
	} {
		service := &MetadataService{
			Root:         tt.root,
			Client:       &test.HttpClient{Resources: tt.resources, Err: tt.clientErr},
			UserdataPath: tt.userdataPath,
		}
		data, err := service.FetchUserdata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(data, tt.userdata) {
			t.Fatalf("bad userdata (%q): want %q, got %q", tt.resources, tt.userdata, data)
		}
	}
}

func TestUrls(t *testing.T) {
	for _, tt := range []struct {
		root         string
		userdataPath string
		metadataPath string
		expectRoot   string
		userdata     string
		metadata     string
	}{
		{
			root:         "/",
			userdataPath: "2009-04-04/user-data",
			metadataPath: "2009-04-04/meta-data",
			expectRoot:   "/",
			userdata:     "/2009-04-04/user-data",
			metadata:     "/2009-04-04/meta-data",
		},
		{
			root:         "http://169.254.169.254/",
			userdataPath: "2009-04-04/user-data",
			metadataPath: "2009-04-04/meta-data",
			expectRoot:   "http://169.254.169.254/",
			userdata:     "http://169.254.169.254/2009-04-04/user-data",
			metadata:     "http://169.254.169.254/2009-04-04/meta-data",
		},
	} {
		service := &MetadataService{
			Root:         tt.root,
			UserdataPath: tt.userdataPath,
			MetadataPath: tt.metadataPath,
		}
		if url := service.UserdataUrl(); url != tt.userdata {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.userdata, url)
		}
		if url := service.MetadataUrl(); url != tt.metadata {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.metadata, url)
		}
		if url := service.ConfigRoot(); url != tt.expectRoot {
			t.Fatalf("bad url (%q): want %q, got %q", tt.root, tt.expectRoot, url)
		}
	}
}

func TestNewDatasource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "/",
		},
		{
			root:       "/",
			expectRoot: "/",
		},
		{
			root:       "http://169.254.169.254",
			expectRoot: "http://169.254.169.254/",
		},
		{
			root:       "http://169.254.169.254/",
			expectRoot: "http://169.254.169.254/",
		},
	} {
		service := NewDatasource(tt.root, "", "", "")
		if service.Root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.Root)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
