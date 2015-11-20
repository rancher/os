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

func TestChild(t *testing.T) {
	tests := []struct {
		parent node
		name   string

		child node
	}{
		{},
		{
			name: "c1",
		},
		{
			parent: node{
				children: []node{
					node{name: "c1"},
					node{name: "c2"},
					node{name: "c3"},
				},
			},
		},
		{
			parent: node{
				children: []node{
					node{name: "c1"},
					node{name: "c2"},
					node{name: "c3"},
				},
			},
			name:  "c2",
			child: node{name: "c2"},
		},
	}

	for _, tt := range tests {
		if child := tt.parent.Child(tt.name); !reflect.DeepEqual(tt.child, child) {
			t.Errorf("bad child (%q): want %#v, got %#v", tt.name, tt.child, child)
		}
	}
}

func TestHumanType(t *testing.T) {
	tests := []struct {
		node node

		humanType string
	}{
		{
			humanType: "invalid",
		},
		{
			node:      node{Value: reflect.ValueOf("hello")},
			humanType: "string",
		},
		{
			node: node{
				Value: reflect.ValueOf([]int{1, 2}),
				children: []node{
					node{Value: reflect.ValueOf(1)},
					node{Value: reflect.ValueOf(2)},
				}},
			humanType: "[]int",
		},
	}

	for _, tt := range tests {
		if humanType := tt.node.HumanType(); tt.humanType != humanType {
			t.Errorf("bad type (%q): want %q, got %q", tt.node, tt.humanType, humanType)
		}
	}
}

func TestToNode(t *testing.T) {
	tests := []struct {
		value   interface{}
		context context

		node node
	}{
		{},
		{
			value: struct{}{},
			node:  node{Value: reflect.ValueOf(struct{}{})},
		},
		{
			value: struct {
				A int `yaml:"a"`
			}{},
			node: node{
				children: []node{
					node{
						name: "a",
						field: reflect.TypeOf(struct {
							A int `yaml:"a"`
						}{}).Field(0),
					},
				},
			},
		},
		{
			value: struct {
				A []int `yaml:"a"`
			}{},
			node: node{
				children: []node{
					node{
						name: "a",
						field: reflect.TypeOf(struct {
							A []int `yaml:"a"`
						}{}).Field(0),
					},
				},
			},
		},
		{
			value: map[interface{}]interface{}{
				"a": map[interface{}]interface{}{
					"b": 2,
				},
			},
			context: NewContext([]byte("a:\n  b: 2")),
			node: node{
				children: []node{
					node{
						line: 1,
						name: "a",
						children: []node{
							node{name: "b", line: 2},
						},
					},
				},
			},
		},
		{
			value: struct {
				A struct {
					Jon bool `yaml:"b"`
				} `yaml:"a"`
			}{},
			node: node{
				children: []node{
					node{
						name: "a",
						children: []node{
							node{
								name: "b",
								field: reflect.TypeOf(struct {
									Jon bool `yaml:"b"`
								}{}).Field(0),
								Value: reflect.ValueOf(false),
							},
						},
						field: reflect.TypeOf(struct {
							A struct {
								Jon bool `yaml:"b"`
							} `yaml:"a"`
						}{}).Field(0),
						Value: reflect.ValueOf(struct {
							Jon bool `yaml:"b"`
						}{}),
					},
				},
				Value: reflect.ValueOf(struct {
					A struct {
						Jon bool `yaml:"b"`
					} `yaml:"a"`
				}{}),
			},
		},
	}

	for _, tt := range tests {
		var node node
		toNode(tt.value, tt.context, &node)
		if !nodesEqual(tt.node, node) {
			t.Errorf("bad node (%#v): want %#v, got %#v", tt.value, tt.node, node)
		}
	}
}

func TestFindKey(t *testing.T) {
	tests := []struct {
		key     string
		context context

		found bool
	}{
		{},
		{
			key:     "key1",
			context: NewContext([]byte("key1: hi")),
			found:   true,
		},
		{
			key:     "key2",
			context: NewContext([]byte("key1: hi")),
			found:   false,
		},
		{
			key:     "key3",
			context: NewContext([]byte("key1:\n  key2:\n    key3: hi")),
			found:   true,
		},
		{
			key:     "key4",
			context: NewContext([]byte("key1:\n  - key4: hi")),
			found:   true,
		},
		{
			key:     "key5",
			context: NewContext([]byte("#key5")),
			found:   false,
		},
	}

	for _, tt := range tests {
		if _, found := findKey(tt.key, tt.context); tt.found != found {
			t.Errorf("bad find (%q): want %t, got %t", tt.key, tt.found, found)
		}
	}
}

func TestFindElem(t *testing.T) {
	tests := []struct {
		context context

		found bool
	}{
		{},
		{
			context: NewContext([]byte("test: hi")),
			found:   false,
		},
		{
			context: NewContext([]byte("test:\n  - a\n  -b")),
			found:   true,
		},
		{
			context: NewContext([]byte("test:\n  -\n    a")),
			found:   true,
		},
	}

	for _, tt := range tests {
		if _, found := findElem(tt.context); tt.found != found {
			t.Errorf("bad find (%q): want %t, got %t", tt.context, tt.found, found)
		}
	}
}

func nodesEqual(a, b node) bool {
	if a.name != b.name ||
		a.line != b.line ||
		!reflect.DeepEqual(a.field, b.field) ||
		len(a.children) != len(b.children) {
		return false
	}
	for i := 0; i < len(a.children); i++ {
		if !nodesEqual(a.children[i], b.children[i]) {
			return false
		}
	}
	return true
}
