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
	"net/url"
	"path"
	"reflect"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"
)

type rule func(config node, report *Report)

// Rules contains all of the validation rules.
var Rules []rule = []rule{
	checkDiscoveryUrl,
	checkEncoding,
	checkStructure,
	checkValidity,
	checkWriteFiles,
	checkWriteFilesUnderCoreos,
}

// checkDiscoveryUrl verifies that the string is a valid url.
func checkDiscoveryUrl(cfg node, report *Report) {
	c := cfg.Child("coreos").Child("etcd").Child("discovery")
	if !c.IsValid() {
		return
	}

	if _, err := url.ParseRequestURI(c.String()); err != nil {
		report.Warning(c.line, "discovery URL is not valid")
	}
}

// checkEncoding validates that, for each file under 'write_files', the
// content can be decoded given the specified encoding.
func checkEncoding(cfg node, report *Report) {
	for _, f := range cfg.Child("write_files").children {
		e := f.Child("encoding")
		if !e.IsValid() {
			continue
		}

		c := f.Child("content")
		if _, err := config.DecodeContent(c.String(), e.String()); err != nil {
			report.Error(c.line, fmt.Sprintf("content cannot be decoded as %q", e.String()))
		}
	}
}

// checkStructure compares the provided config to the empty config.CloudConfig
// structure. Each node is checked to make sure that it exists in the known
// structure and that its type is compatible.
func checkStructure(cfg node, report *Report) {
	g := NewNode(config.CloudConfig{}, NewContext([]byte{}))
	checkNodeStructure(cfg, g, report)
}

func checkNodeStructure(n, g node, r *Report) {
	if !isCompatible(n.Kind(), g.Kind()) {
		r.Warning(n.line, fmt.Sprintf("incorrect type for %q (want %s)", n.name, g.HumanType()))
		return
	}

	switch g.Kind() {
	case reflect.Struct:
		for _, cn := range n.children {
			if cg := g.Child(cn.name); cg.IsValid() {
				if msg := cg.field.Tag.Get("deprecated"); msg != "" {
					r.Warning(cn.line, fmt.Sprintf("deprecated key %q (%s)", cn.name, msg))
				}
				checkNodeStructure(cn, cg, r)
			} else {
				r.Warning(cn.line, fmt.Sprintf("unrecognized key %q", cn.name))
			}
		}
	case reflect.Slice:
		for _, cn := range n.children {
			var cg node
			c := g.Type().Elem()
			toNode(reflect.New(c).Elem().Interface(), context{}, &cg)
			checkNodeStructure(cn, cg, r)
		}
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
	default:
		panic(fmt.Sprintf("checkNodeStructure(): unhandled kind %s", g.Kind()))
	}
}

// isCompatible determines if the type of kind n can be converted to the type
// of kind g in the context of YAML. This is not an exhaustive list, but its
// enough for the purposes of cloud-config validation.
func isCompatible(n, g reflect.Kind) bool {
	switch g {
	case reflect.String:
		return n == reflect.String || n == reflect.Int || n == reflect.Float64 || n == reflect.Bool
	case reflect.Struct:
		return n == reflect.Struct || n == reflect.Map
	case reflect.Float64:
		return n == reflect.Float64 || n == reflect.Int
	case reflect.Bool, reflect.Slice, reflect.Int:
		return n == g
	default:
		panic(fmt.Sprintf("isCompatible(): unhandled kind %s", g))
	}
}

// checkValidity checks the value of every node in the provided config by
// running config.AssertValid() on it.
func checkValidity(cfg node, report *Report) {
	g := NewNode(config.CloudConfig{}, NewContext([]byte{}))
	checkNodeValidity(cfg, g, report)
}

func checkNodeValidity(n, g node, r *Report) {
	if err := config.AssertValid(n.Value, g.field.Tag.Get("valid")); err != nil {
		r.Error(n.line, fmt.Sprintf("invalid value %v", n.Value.Interface()))
	}
	switch g.Kind() {
	case reflect.Struct:
		for _, cn := range n.children {
			if cg := g.Child(cn.name); cg.IsValid() {
				checkNodeValidity(cn, cg, r)
			}
		}
	case reflect.Slice:
		for _, cn := range n.children {
			var cg node
			c := g.Type().Elem()
			toNode(reflect.New(c).Elem().Interface(), context{}, &cg)
			checkNodeValidity(cn, cg, r)
		}
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
	default:
		panic(fmt.Sprintf("checkNodeValidity(): unhandled kind %s", g.Kind()))
	}
}

// checkWriteFiles checks to make sure that the target file can actually be
// written. Note that this check is approximate (it only checks to see if the file
// is under /usr).
func checkWriteFiles(cfg node, report *Report) {
	for _, f := range cfg.Child("write_files").children {
		c := f.Child("path")
		if !c.IsValid() {
			continue
		}

		d := path.Dir(c.String())
		switch {
		case strings.HasPrefix(d, "/usr"):
			report.Error(c.line, "file cannot be written to a read-only filesystem")
		}
	}
}

// checkWriteFilesUnderCoreos checks to see if the 'write_files' node is a
// child of 'coreos' (it shouldn't be).
func checkWriteFilesUnderCoreos(cfg node, report *Report) {
	c := cfg.Child("coreos").Child("write_files")
	if c.IsValid() {
		report.Info(c.line, "write_files doesn't belong under coreos")
	}
}
