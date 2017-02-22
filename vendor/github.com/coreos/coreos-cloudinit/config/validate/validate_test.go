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

package validate

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseCloudConfig(t *testing.T) {
	tests := []struct {
		config string

		entries []Entry
	}{
		{},
		{
			config: "	",
			entries: []Entry{{entryError, "found character that cannot start any token", 1}},
		},
		{
			config:  "a:\na",
			entries: []Entry{{entryError, "could not find expected ':'", 2}},
		},
		{
			config:  "#hello\na:\na",
			entries: []Entry{{entryError, "could not find expected ':'", 3}},
		},
	}

	for _, tt := range tests {
		r := Report{}
		parseCloudConfig([]byte(tt.config), &r)

		if e := r.Entries(); !reflect.DeepEqual(tt.entries, e) {
			t.Errorf("bad report (%s): want %#v, got %#v", tt.config, tt.entries, e)
		}
	}
}

func TestValidateCloudConfig(t *testing.T) {
	tests := []struct {
		config string
		rules  []rule

		report Report
		err    error
	}{
		{
			rules: []rule{func(_ node, _ *Report) { panic("something happened") }},
			err:   errors.New("something happened"),
		},
		{
			config: "write_files:\n  - permissions: 0744",
			rules:  Rules,
		},
		{
			config: "write_files:\n  - permissions: '0744'",
			rules:  Rules,
		},
		{
			config: "write_files:\n  - permissions: 744",
			rules:  Rules,
		},
		{
			config: "write_files:\n  - permissions: '744'",
			rules:  Rules,
		},
		{
			config: "coreos:\n  update:\n    reboot-strategy: off",
			rules:  Rules,
		},
		{
			config: "coreos:\n  update:\n    reboot-strategy: false",
			rules:  Rules,
			report: Report{entries: []Entry{{entryError, "invalid value false", 3}}},
		},
	}

	for _, tt := range tests {
		r, err := validateCloudConfig([]byte(tt.config), tt.rules)
		if !reflect.DeepEqual(tt.err, err) {
			t.Errorf("bad error (%s): want %v, got %v", tt.config, tt.err, err)
		}
		if !reflect.DeepEqual(tt.report, r) {
			t.Errorf("bad report (%s): want %+v, got %+v", tt.config, tt.report, r)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		config string

		report Report
	}{
		{},
		{
			config: "#!/bin/bash\necho hey",
		},
		{
			config: "{}",
			report: Report{entries: []Entry{{entryError, `must be "#cloud-config" or begin with "#!"`, 1}}},
		},
		{
			config: `{"ignitionVersion":0}`,
		},
		{
			config: `{"ignitionVersion":1}`,
		},
	}

	for i, tt := range tests {
		r, err := Validate([]byte(tt.config))
		if err != nil {
			t.Errorf("bad error (case #%d): want %v, got %v", i, nil, err)
		}
		if !reflect.DeepEqual(tt.report, r) {
			t.Errorf("bad report (case #%d): want %+v, got %+v", i, tt.report, r)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	config := `#cloud-config
hostname: test

coreos:
  etcd:
    name:      node001
    discovery: https://discovery.etcd.io/disco
    addr:      $public_ipv4:4001
    peer-addr: $private_ipv4:7001
  fleet:
    verbosity: 2
    metadata:  "hi"
  update:
    reboot-strategy: off
  units:
    - name:    hi.service
      command: start
      enable:  true
    - name:    bye.service
      command: stop

ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC0g+ZTxC7weoIJLUafOgrm+h...

users:
  - name: me

write_files:
  - path: /etc/yes
    content: "Hi"

manage_etc_hosts: localhost`

	for i := 0; i < b.N; i++ {
		if _, err := Validate([]byte(config)); err != nil {
			panic(err)
		}
	}
}
