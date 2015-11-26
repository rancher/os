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

package network

import (
	"testing"
)

func TestFormatConfigs(t *testing.T) {
	for in, n := range map[string]int{
		"":                                                    0,
		"line1\\\nis long":                                    1,
		"#comment":                                            0,
		"#comment\\\ncomment":                                 0,
		"  #comment \\\n comment\nline 1\nline 2\\\n is long": 2,
	} {
		lines := formatConfig(in)
		if len(lines) != n {
			t.Fatalf("bad number of lines for config %q: got %d, want %d", in, len(lines), n)
		}
	}
}

func TestProcessDebianNetconf(t *testing.T) {
	for _, tt := range []struct {
		in   string
		fail bool
		n    int
	}{
		{"", false, 0},
		{"iface", true, -1},
		{"auto eth1\nauto eth2", false, 0},
		{"iface eth1 inet manual", false, 1},
	} {
		interfaces, err := ProcessDebianNetconf([]byte(tt.in))
		failed := err != nil
		if tt.fail != failed {
			t.Fatalf("bad failure state for %q: got %t, want %t", tt.in, failed, tt.fail)
		}
		if tt.n != -1 && tt.n != len(interfaces) {
			t.Fatalf("bad number of interfaces for %q: got %d, want %q", tt.in, len(interfaces), tt.n)
		}
	}
}
