/*
   Copyright 2014 CoreOS, Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"testing"
)

func TestRebootStrategyValid(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "best-effort", isValid: true},
		{value: "etcd-lock", isValid: true},
		{value: "reboot", isValid: true},
		{value: "off", isValid: true},
		{value: "besteffort", isValid: false},
		{value: "unknown", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(Update{RebootStrategy: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}
