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

package config

import (
	"testing"
)

func TestRebootWindowStart(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "Sun 0:0", isValid: true},
		{value: "Sun 00:00", isValid: true},
		{value: "sUn 23:59", isValid: true},
		{value: "mon 0:0", isValid: true},
		{value: "tue 0:0", isValid: true},
		{value: "tues 0:0", isValid: false},
		{value: "wed 0:0", isValid: true},
		{value: "thu 0:0", isValid: true},
		{value: "thur 0:0", isValid: false},
		{value: "fri 0:0", isValid: true},
		{value: "sat 0:0", isValid: true},
		{value: "sat00:00", isValid: false},
		{value: "00:00", isValid: true},
		{value: "10:10", isValid: true},
		{value: "20:20", isValid: true},
		{value: "20:30", isValid: true},
		{value: "20:40", isValid: true},
		{value: "20:50", isValid: true},
		{value: "20:60", isValid: false},
		{value: "24:00", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(Locksmith{RebootWindowStart: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}

func TestRebootWindowLength(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "1h", isValid: true},
		{value: "1d", isValid: true},
		{value: "0d", isValid: true},
		{value: "0.5h", isValid: true},
		{value: "0.5.0h", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(Locksmith{RebootWindowLength: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}
