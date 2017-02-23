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
	"fmt"
	"reflect"
	"regexp"
)

var (
	yamlKey  = regexp.MustCompile(`^ *-? ?(?P<key>.*?):`)
	yamlElem = regexp.MustCompile(`^ *-`)
)

type Node struct {
	name     string
	line     int
	children []Node
	field    reflect.StructField
	reflect.Value
}

// Child attempts to find the child with the given name in the Node's list of
// children. If no such child is found, an invalid Node is returned.
func (n Node) Child(name string) Node {
	for _, c := range n.children {
		if c.name == name {
			return c
		}
	}
	return Node{}
}

// HumanType returns the human-consumable string representation of the type of
// the Node.
func (n Node) HumanType() string {
	switch k := n.Kind(); k {
	case reflect.Slice:
		c := n.Type().Elem()
		return "[]" + Node{Value: reflect.New(c).Elem()}.HumanType()
	default:
		return k.String()
	}
}

// NewNode returns the Node representation of the given value. The context
// will be used in an attempt to determine line numbers for the given value.
func NewNode(value interface{}, context Context) Node {
	var n Node
	toNode(value, context, &n)
	return n
}

// toNode converts the given value into a Node and then recursively processes
// each of the Nodes components (e.g. fields, array elements, keys).
func toNode(v interface{}, c Context, n *Node) {
	vv := reflect.ValueOf(v)
	if !vv.IsValid() {
		return
	}

	n.Value = vv
	switch vv.Kind() {
	case reflect.Struct:
		// Walk over each field in the structure, skipping unexported fields,
		// and create a Node for it.
		for i := 0; i < vv.Type().NumField(); i++ {
			ft := vv.Type().Field(i)
			k := ft.Tag.Get("yaml")
			if k == "-" || k == "" {
				continue
			}

			cn := Node{name: k, field: ft}
			c, ok := findKey(cn.name, c)
			if ok {
				cn.line = c.lineNumber
			}
			toNode(vv.Field(i).Interface(), c, &cn)
			n.children = append(n.children, cn)
		}
	case reflect.Map:
		// Walk over each key in the map and create a Node for it.
		v := v.(map[interface{}]interface{})
		for k, cv := range v {
			cn := Node{name: fmt.Sprintf("%s", k)}
			c, ok := findKey(cn.name, c)
			if ok {
				cn.line = c.lineNumber
			}
			toNode(cv, c, &cn)
			n.children = append(n.children, cn)
		}
	case reflect.Slice:
		// Walk over each element in the slice and create a Node for it.
		// While iterating over the slice, preserve the context after it
		// is modified. This allows the line numbers to reflect the current
		// element instead of the first.
		for i := 0; i < vv.Len(); i++ {
			cn := Node{
				name:  fmt.Sprintf("%s[%d]", n.name, i),
				field: n.field,
			}
			var ok bool
			c, ok = findElem(c)
			if ok {
				cn.line = c.lineNumber
			}
			toNode(vv.Index(i).Interface(), c, &cn)
			n.children = append(n.children, cn)
			c.Increment()
		}
	case reflect.String, reflect.Int, reflect.Bool, reflect.Float64:
	default:
		panic(fmt.Sprintf("toNode(): unhandled kind %s", vv.Kind()))
	}
}

// findKey attempts to find the requested key within the provided context.
// A modified copy of the context is returned with every line up to the key
// incremented past. A boolean, true if the key was found, is also returned.
func findKey(key string, context Context) (Context, bool) {
	return find(yamlKey, key, context)
}

// findElem attempts to find an array element within the provided context.
// A modified copy of the context is returned with every line up to the array
// element incremented past. A boolean, true if the key was found, is also
// returned.
func findElem(context Context) (Context, bool) {
	return find(yamlElem, "", context)
}

func find(exp *regexp.Regexp, key string, context Context) (Context, bool) {
	for len(context.currentLine) > 0 || len(context.remainingLines) > 0 {
		matches := exp.FindStringSubmatch(context.currentLine)
		if len(matches) > 0 && (key == "" || matches[1] == key) {
			return context, true
		}

		context.Increment()
	}
	return context, false
}
