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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/coreos/coreos-cloudinit/config"

	"github.com/coreos/yaml"
)

var (
	yamlLineError = regexp.MustCompile(`^YAML error: line (?P<line>[[:digit:]]+): (?P<msg>.*)$`)
	yamlError     = regexp.MustCompile(`^YAML error: (?P<msg>.*)$`)
)

// Validate runs a series of validation tests against the given userdata and
// returns a report detailing all of the issues. Presently, only cloud-configs
// can be validated.
func Validate(userdataBytes []byte) (Report, error) {
	switch {
	case len(userdataBytes) == 0:
		return Report{}, nil
	case config.IsScript(string(userdataBytes)):
		return Report{}, nil
	case config.IsIgnitionConfig(string(userdataBytes)):
		return Report{}, nil
	case config.IsCloudConfig(string(userdataBytes)):
		return validateCloudConfig(userdataBytes, Rules)
	default:
		return Report{entries: []Entry{
			{kind: entryError, message: `must be "#cloud-config" or begin with "#!"`, line: 1},
		}}, nil
	}
}

// validateCloudConfig runs all of the validation rules in Rules and returns
// the resulting report and any errors encountered.
func validateCloudConfig(config []byte, rules []rule) (report Report, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	c, err := parseCloudConfig(config, &report)
	if err != nil {
		return report, err
	}

	for _, r := range rules {
		r(c, &report)
	}
	return report, nil
}

// parseCloudConfig parses the provided config into a node structure and logs
// any parsing issues into the provided report. Unrecoverable errors are
// returned as an error.
func parseCloudConfig(cfg []byte, report *Report) (node, error) {
	yaml.UnmarshalMappingKeyTransform = func(nameIn string) (nameOut string) {
		return nameIn
	}
	// unmarshal the config into an implicitly-typed form. The yaml library
	// will implicitly convert types into their normalized form
	// (e.g. 0744 -> 484, off -> false).
	var weak map[interface{}]interface{}
	if err := yaml.Unmarshal(cfg, &weak); err != nil {
		matches := yamlLineError.FindStringSubmatch(err.Error())
		if len(matches) == 3 {
			line, err := strconv.Atoi(matches[1])
			if err != nil {
				return node{}, err
			}
			msg := matches[2]
			report.Error(line, msg)
			return node{}, nil
		}

		matches = yamlError.FindStringSubmatch(err.Error())
		if len(matches) == 2 {
			report.Error(1, matches[1])
			return node{}, nil
		}

		return node{}, errors.New("couldn't parse yaml error")
	}
	w := NewNode(weak, NewContext(cfg))
	w = normalizeNodeNames(w, report)

	// unmarshal the config into the explicitly-typed form.
	yaml.UnmarshalMappingKeyTransform = func(nameIn string) (nameOut string) {
		return strings.Replace(nameIn, "-", "_", -1)
	}
	var strong config.CloudConfig
	if err := yaml.Unmarshal([]byte(cfg), &strong); err != nil {
		return node{}, err
	}
	s := NewNode(strong, NewContext(cfg))

	// coerceNodes weak nodes and strong nodes. strong nodes replace weak nodes
	// if they are compatible types (this happens when the yaml library
	// converts the input).
	// (e.g. weak 484 is replaced by strong 0744, weak 4 is not replaced by
	// strong false)
	return coerceNodes(w, s), nil
}

// coerceNodes recursively evaluates two nodes, returning a new node containing
// either the weak or strong node's value and its recursively processed
// children. The strong node's value is used if the two nodes are leafs, are
// both valid, and are compatible types (defined by isCompatible()). The weak
// node is returned in all other cases. coerceNodes is used to counteract the
// effects of yaml's automatic type conversion. The weak node is the one
// resulting from unmarshalling into an empty interface{} (the type is
// inferred). The strong node is the one resulting from unmarshalling into a
// struct. If the two nodes are of compatible types, the yaml library correctly
// parsed the value into the strongly typed unmarshalling. In this case, we
// prefer the strong node because its actually the type we are expecting.
func coerceNodes(w, s node) node {
	n := w
	n.children = nil
	if len(w.children) == 0 && len(s.children) == 0 &&
		w.IsValid() && s.IsValid() &&
		isCompatible(w.Kind(), s.Kind()) {
		n.Value = s.Value
	}

	for _, cw := range w.children {
		n.children = append(n.children, coerceNodes(cw, s.Child(cw.name)))
	}
	return n
}

// normalizeNodeNames replaces all occurences of '-' with '_' within key names
// and makes a note of each replacement in the report.
func normalizeNodeNames(node node, report *Report) node {
	if strings.Contains(node.name, "-") {
		// TODO(crawford): Enable this message once the new validator hits stable.
		//report.Info(node.line, fmt.Sprintf("%q uses '-' instead of '_'", node.name))
		node.name = strings.Replace(node.name, "-", "_", -1)
	}
	for i := range node.children {
		node.children[i] = normalizeNodeNames(node.children[i], report)
	}
	return node
}
