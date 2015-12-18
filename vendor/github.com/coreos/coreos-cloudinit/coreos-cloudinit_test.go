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

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"reflect"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/datasource"
)

func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		cc *config.CloudConfig
		md datasource.Metadata

		out config.CloudConfig
	}{
		{
			// If md is empty and cc is nil, result should be empty
			out: config.CloudConfig{},
		},
		{
			// If md and cc are empty, result should be empty
			cc:  &config.CloudConfig{},
			out: config.CloudConfig{},
		},
		{
			// If cc is empty, cc should be returned unchanged
			cc:  &config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def"}, Hostname: "cc-host"},
			out: config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def"}, Hostname: "cc-host"},
		},
		{
			// If cc is empty, cc should be returned unchanged(overridden)
			cc:  &config.CloudConfig{},
			md:  datasource.Metadata{Hostname: "md-host", SSHPublicKeys: map[string]string{"key": "ghi"}},
			out: config.CloudConfig{SSHAuthorizedKeys: []string{"ghi"}, Hostname: "md-host"},
		},
		{
			// If cc is nil, cc should be returned unchanged(overridden)
			md:  datasource.Metadata{Hostname: "md-host", SSHPublicKeys: map[string]string{"key": "ghi"}},
			out: config.CloudConfig{SSHAuthorizedKeys: []string{"ghi"}, Hostname: "md-host"},
		},
		{
			// user-data should override completely in the case of conflicts
			cc:  &config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def"}, Hostname: "cc-host"},
			md:  datasource.Metadata{Hostname: "md-host"},
			out: config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def"}, Hostname: "cc-host"},
		},
		{
			// Mixed merge should succeed
			cc:  &config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def"}, Hostname: "cc-host"},
			md:  datasource.Metadata{Hostname: "md-host", SSHPublicKeys: map[string]string{"key": "ghi"}},
			out: config.CloudConfig{SSHAuthorizedKeys: []string{"abc", "def", "ghi"}, Hostname: "cc-host"},
		},
		{
			// Completely non-conflicting merge should be fine
			cc:  &config.CloudConfig{Hostname: "cc-host"},
			md:  datasource.Metadata{SSHPublicKeys: map[string]string{"zaphod": "beeblebrox"}},
			out: config.CloudConfig{Hostname: "cc-host", SSHAuthorizedKeys: []string{"beeblebrox"}},
		},
		{
			// Non-mergeable settings in user-data should not be affected
			cc:  &config.CloudConfig{Hostname: "cc-host", ManageEtcHosts: config.EtcHosts("lolz")},
			md:  datasource.Metadata{Hostname: "md-host"},
			out: config.CloudConfig{Hostname: "cc-host", ManageEtcHosts: config.EtcHosts("lolz")},
		},
	}

	for i, tt := range tests {
		out := mergeConfigs(tt.cc, tt.md)
		if !reflect.DeepEqual(tt.out, out) {
			t.Errorf("bad config (%d): want %#v, got %#v", i, tt.out, out)
		}
	}
}

func mustDecode(in string) []byte {
	out, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}
	return out
}

func TestDecompressIfGzip(t *testing.T) {
	tests := []struct {
		in []byte

		out []byte
		err error
	}{
		{
			in: nil,

			out: nil,
			err: nil,
		},
		{
			in: []byte{},

			out: []byte{},
			err: nil,
		},
		{
			in: mustDecode("H4sIAJWV/VUAA1NOzskvTdFNzs9Ly0wHABt6mQENAAAA"),

			out: []byte("#cloud-config"),
			err: nil,
		},
		{
			in: []byte("#cloud-config"),

			out: []byte("#cloud-config"),
			err: nil,
		},
		{
			in: mustDecode("H4sCORRUPT=="),

			out: nil,
			err: errors.New("any error"),
		},
	}
	for i, tt := range tests {
		out, err := decompressIfGzip(tt.in)
		if !bytes.Equal(out, tt.out) || (tt.err != nil && err == nil) {
			t.Errorf("bad gzip (%d): want (%s, %#v), got (%s, %#v)", i, string(tt.out), tt.err, string(out), err)
		}
	}

}
