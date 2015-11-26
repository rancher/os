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
	"bytes"
	"reflect"
	"testing"
)

func TestEntry(t *testing.T) {
	tests := []struct {
		entry Entry

		str  string
		json []byte
	}{
		{
			Entry{entryInfo, "test info", 1},
			"line 1: info: test info",
			[]byte(`{"kind":"info","line":1,"message":"test info"}`),
		},
		{
			Entry{entryWarning, "test warning", 1},
			"line 1: warning: test warning",
			[]byte(`{"kind":"warning","line":1,"message":"test warning"}`),
		},
		{
			Entry{entryError, "test error", 2},
			"line 2: error: test error",
			[]byte(`{"kind":"error","line":2,"message":"test error"}`),
		},
	}

	for _, tt := range tests {
		if str := tt.entry.String(); tt.str != str {
			t.Errorf("bad string (%q): want %q, got %q", tt.entry, tt.str, str)
		}
		json, err := tt.entry.MarshalJSON()
		if err != nil {
			t.Errorf("bad error (%q): want %v, got %q", tt.entry, nil, err)
		}
		if !bytes.Equal(tt.json, json) {
			t.Errorf("bad JSON (%q): want %q, got %q", tt.entry, tt.json, json)
		}
	}
}

func TestReport(t *testing.T) {
	type reportFunc struct {
		fn      func(*Report, int, string)
		line    int
		message string
	}

	tests := []struct {
		fs []reportFunc

		es []Entry
	}{
		{
			[]reportFunc{
				{(*Report).Warning, 1, "test warning 1"},
				{(*Report).Error, 2, "test error 2"},
				{(*Report).Info, 10, "test info 10"},
			},
			[]Entry{
				Entry{entryWarning, "test warning 1", 1},
				Entry{entryError, "test error 2", 2},
				Entry{entryInfo, "test info 10", 10},
			},
		},
	}

	for _, tt := range tests {
		r := Report{}
		for _, f := range tt.fs {
			f.fn(&r, f.line, f.message)
		}
		if es := r.Entries(); !reflect.DeepEqual(tt.es, es) {
			t.Errorf("bad entries (%v): want %#v, got %#v", tt.fs, tt.es, es)
		}
	}
}
