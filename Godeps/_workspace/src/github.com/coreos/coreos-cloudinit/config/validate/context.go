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
	"strings"
)

// context represents the current position within a newline-delimited string.
// Each line is loaded, one by one, into currentLine (newline omitted) and
// lineNumber keeps track of its position within the original string.
type context struct {
	currentLine    string
	remainingLines string
	lineNumber     int
}

// Increment moves the context to the next line (if available).
func (c *context) Increment() {
	if c.currentLine == "" && c.remainingLines == "" {
		return
	}

	lines := strings.SplitN(c.remainingLines, "\n", 2)
	c.currentLine = lines[0]
	if len(lines) == 2 {
		c.remainingLines = lines[1]
	} else {
		c.remainingLines = ""
	}
	c.lineNumber++
}

// NewContext creates a context from the provided data. It strips out all
// carriage returns and moves to the first line (if available).
func NewContext(content []byte) context {
	c := context{remainingLines: strings.Replace(string(content), "\r", "", -1)}
	c.Increment()
	return c
}
