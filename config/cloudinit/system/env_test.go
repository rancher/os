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

package system

import (
	"testing"
)

func TestServiceContents(t *testing.T) {
	tests := []struct {
		Config   interface{}
		Contents string
	}{
		{
			struct{}{},
			"",
		},
		{
			struct {
				A string  `env:"A"`
				B int     `env:"B"`
				C bool    `env:"C"`
				D float64 `env:"D"`
			}{
				"hi", 1, true, 0.12345,
			},
			`[Service]
Environment="A=hi"
Environment="B=1"
Environment="C=true"
Environment="D=0.12345"
`,
		},
		{
			struct {
				A float64 `env:"A"`
				B float64 `env:"B"`
				C float64 `env:"C"`
				D float64 `env:"D"`
			}{
				0.000001, 1, 0.9999999, 0.1,
			},
			`[Service]
Environment="A=1e-06"
Environment="B=1"
Environment="C=0.9999999"
Environment="D=0.1"
`,
		},
	}

	for _, tt := range tests {
		if c := serviceContents(tt.Config); c != tt.Contents {
			t.Errorf("bad contents (%+v): want %q, got %q", tt, tt.Contents, c)
		}
	}
}
