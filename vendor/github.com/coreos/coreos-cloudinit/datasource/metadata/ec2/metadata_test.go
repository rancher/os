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

package ec2

import (
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/metadata"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/test"
	"github.com/coreos/coreos-cloudinit/pkg"
)

func TestType(t *testing.T) {
	want := "ec2-metadata-service"
	if kind := (metadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestFetchAttributes(t *testing.T) {
	for _, s := range []struct {
		resources map[string]string
		err       error
		tests     []struct {
			path string
			val  []string
		}
	}{
		{
			resources: map[string]string{
				"/":      "a\nb\nc/",
				"/c/":    "d\ne/",
				"/c/e/":  "f",
				"/a":     "1",
				"/b":     "2",
				"/c/d":   "3",
				"/c/e/f": "4",
			},
			tests: []struct {
				path string
				val  []string
			}{
				{"/", []string{"a", "b", "c/"}},
				{"/b", []string{"2"}},
				{"/c/d", []string{"3"}},
				{"/c/e/", []string{"f"}},
			},
		},
		{
			err: fmt.Errorf("test error"),
			tests: []struct {
				path string
				val  []string
			}{
				{"", nil},
			},
		},
	} {
		service := metadataService{metadata.MetadataService{
			Client: &test.HttpClient{Resources: s.resources, Err: s.err},
		}}
		for _, tt := range s.tests {
			attrs, err := service.fetchAttributes(tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.resources, s.err, err)
			}
			if !reflect.DeepEqual(attrs, tt.val) {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.resources, tt.val, attrs)
			}
		}
	}
}

func TestFetchAttribute(t *testing.T) {
	for _, s := range []struct {
		resources map[string]string
		err       error
		tests     []struct {
			path string
			val  string
		}
	}{
		{
			resources: map[string]string{
				"/":      "a\nb\nc/",
				"/c/":    "d\ne/",
				"/c/e/":  "f",
				"/a":     "1",
				"/b":     "2",
				"/c/d":   "3",
				"/c/e/f": "4",
			},
			tests: []struct {
				path string
				val  string
			}{
				{"/a", "1"},
				{"/b", "2"},
				{"/c/d", "3"},
				{"/c/e/f", "4"},
			},
		},
		{
			err: fmt.Errorf("test error"),
			tests: []struct {
				path string
				val  string
			}{
				{"", ""},
			},
		},
	} {
		service := metadataService{metadata.MetadataService{
			Client: &test.HttpClient{Resources: s.resources, Err: s.err},
		}}
		for _, tt := range s.tests {
			attr, err := service.fetchAttribute(tt.path)
			if err != s.err {
				t.Fatalf("bad error for %q (%q): want %q, got %q", tt.path, s.resources, s.err, err)
			}
			if attr != tt.val {
				t.Fatalf("bad fetch for %q (%q): want %q, got %q", tt.path, s.resources, tt.val, attr)
			}
		}
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		metadataPath string
		resources    map[string]string
		expect       datasource.Metadata
		clientErr    error
		expectErr    error
	}{
		{
			root:         "/",
			metadataPath: "2009-04-04/meta-data",
			resources: map[string]string{
				"/2009-04-04/meta-data/public-keys": "bad\n",
			},
			expectErr: fmt.Errorf("malformed public key: \"bad\""),
		},
		{
			root:         "/",
			metadataPath: "2009-04-04/meta-data",
			resources: map[string]string{
				"/2009-04-04/meta-data/hostname":                  "host",
				"/2009-04-04/meta-data/local-ipv4":                "1.2.3.4",
				"/2009-04-04/meta-data/public-ipv4":               "5.6.7.8",
				"/2009-04-04/meta-data/public-keys":               "0=test1\n",
				"/2009-04-04/meta-data/public-keys/0":             "openssh-key",
				"/2009-04-04/meta-data/public-keys/0/openssh-key": "key",
			},
			expect: datasource.Metadata{
				Hostname:      "host",
				PrivateIPv4:   net.ParseIP("1.2.3.4"),
				PublicIPv4:    net.ParseIP("5.6.7.8"),
				SSHPublicKeys: map[string]string{"test1": "key"},
			},
		},
		{
			root:         "/",
			metadataPath: "2009-04-04/meta-data",
			resources: map[string]string{
				"/2009-04-04/meta-data/hostname":                  "host domain another_domain",
				"/2009-04-04/meta-data/local-ipv4":                "1.2.3.4",
				"/2009-04-04/meta-data/public-ipv4":               "5.6.7.8",
				"/2009-04-04/meta-data/public-keys":               "0=test1\n",
				"/2009-04-04/meta-data/public-keys/0":             "openssh-key",
				"/2009-04-04/meta-data/public-keys/0/openssh-key": "key",
			},
			expect: datasource.Metadata{
				Hostname:      "host",
				PrivateIPv4:   net.ParseIP("1.2.3.4"),
				PublicIPv4:    net.ParseIP("5.6.7.8"),
				SSHPublicKeys: map[string]string{"test1": "key"},
			},
		},
		{
			clientErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
			expectErr: pkg.ErrTimeout{Err: fmt.Errorf("test error")},
		},
	} {
		service := &metadataService{metadata.MetadataService{
			Root:         tt.root,
			Client:       &test.HttpClient{Resources: tt.resources, Err: tt.clientErr},
			MetadataPath: tt.metadataPath,
		}}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !reflect.DeepEqual(tt.expect, metadata) {
			t.Fatalf("bad fetch (%q): want %#v, got %#v", tt.resources, tt.expect, metadata)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
