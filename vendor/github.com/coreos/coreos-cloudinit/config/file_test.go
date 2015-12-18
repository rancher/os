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

func TestEncodingValid(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "base64", isValid: true},
		{value: "b64", isValid: true},
		{value: "gz", isValid: true},
		{value: "gzip", isValid: true},
		{value: "gz+base64", isValid: true},
		{value: "gzip+base64", isValid: true},
		{value: "gz+b64", isValid: true},
		{value: "gzip+b64", isValid: true},
		{value: "gzzzzbase64", isValid: false},
		{value: "gzipppbase64", isValid: false},
		{value: "unknown", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(File{Encoding: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}

func TestRawFilePermissionsValid(t *testing.T) {
	tests := []struct {
		value string

		isValid bool
	}{
		{value: "744", isValid: true},
		{value: "0744", isValid: true},
		{value: "1744", isValid: true},
		{value: "01744", isValid: true},
		{value: "11744", isValid: false},
		{value: "rwxr--r--", isValid: false},
		{value: "800", isValid: false},
	}

	for _, tt := range tests {
		isValid := (nil == AssertStructValid(File{RawFilePermissions: tt.value}))
		if tt.isValid != isValid {
			t.Errorf("bad assert (%s): want %t, got %t", tt.value, tt.isValid, isValid)
		}
	}
}
