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
	"reflect"
	"testing"
)

func TestNewContext(t *testing.T) {
	tests := []struct {
		in string

		out Context
	}{
		{
			out: Context{
				currentLine:    "",
				remainingLines: "",
				lineNumber:     0,
			},
		},
		{
			in: "this\r\nis\r\na\r\ntest",
			out: Context{
				currentLine:    "this",
				remainingLines: "is\na\ntest",
				lineNumber:     1,
			},
		},
	}

	for _, tt := range tests {
		if out := NewContext([]byte(tt.in)); !reflect.DeepEqual(tt.out, out) {
			t.Errorf("bad context (%q): want %#v, got %#v", tt.in, tt.out, out)
		}
	}
}

func TestIncrement(t *testing.T) {
	tests := []struct {
		init Context
		op   func(c *Context)

		res Context
	}{
		{
			init: Context{
				currentLine:    "",
				remainingLines: "",
				lineNumber:     0,
			},
			res: Context{
				currentLine:    "",
				remainingLines: "",
				lineNumber:     0,
			},
			op: func(c *Context) {
				c.Increment()
			},
		},
		{
			init: Context{
				currentLine:    "test",
				remainingLines: "",
				lineNumber:     1,
			},
			res: Context{
				currentLine:    "",
				remainingLines: "",
				lineNumber:     2,
			},
			op: func(c *Context) {
				c.Increment()
				c.Increment()
				c.Increment()
			},
		},
		{
			init: Context{
				currentLine:    "this",
				remainingLines: "is\na\ntest",
				lineNumber:     1,
			},
			res: Context{
				currentLine:    "is",
				remainingLines: "a\ntest",
				lineNumber:     2,
			},
			op: func(c *Context) {
				c.Increment()
			},
		},
		{
			init: Context{
				currentLine:    "this",
				remainingLines: "is\na\ntest",
				lineNumber:     1,
			},
			res: Context{
				currentLine:    "test",
				remainingLines: "",
				lineNumber:     4,
			},
			op: func(c *Context) {
				c.Increment()
				c.Increment()
				c.Increment()
			},
		},
	}

	for i, tt := range tests {
		res := tt.init
		if tt.op(&res); !reflect.DeepEqual(tt.res, res) {
			t.Errorf("bad context (%d, %#v): want %#v, got %#v", i, tt.init, tt.res, res)
		}
	}
}
