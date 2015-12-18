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

package initialize

import (
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestParseHeaderCRLF(t *testing.T) {
	configs := []string{
		"#cloud-config\nfoo: bar",
		"#cloud-config\r\nfoo: bar",
	}

	for i, config := range configs {
		_, err := ParseUserData(config)
		if err != nil {
			t.Errorf("Failed parsing config %d: %v", i, err)
		}
	}

	scripts := []string{
		"#!bin/bash\necho foo",
		"#!bin/bash\r\necho foo",
	}

	for i, script := range scripts {
		_, err := ParseUserData(script)
		if err != nil {
			t.Errorf("Failed parsing script %d: %v", i, err)
		}
	}
}

func TestParseConfigCRLF(t *testing.T) {
	contents := "#cloud-config \r\nhostname: foo\r\nssh_authorized_keys:\r\n  - foobar\r\n"
	ud, err := ParseUserData(contents)
	if err != nil {
		t.Fatalf("Failed parsing config: %v", err)
	}

	cfg := ud.(*config.CloudConfig)

	if cfg.Hostname != "foo" {
		t.Error("Failed parsing hostname from config")
	}

	if len(cfg.SSHAuthorizedKeys) != 1 {
		t.Error("Parsed incorrect number of SSH keys")
	}
}

func TestParseConfigEmpty(t *testing.T) {
	i, e := ParseUserData(``)
	if i != nil {
		t.Error("ParseUserData of empty string returned non-nil unexpectedly")
	} else if e != nil {
		t.Error("ParseUserData of empty string returned error unexpectedly")
	}
}
