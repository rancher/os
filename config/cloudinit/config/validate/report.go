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
	"encoding/json"
	"fmt"
)

// Report represents the list of entries resulting from validation.
type Report struct {
	entries []Entry
}

// Error adds an error entry to the report.
func (r *Report) Error(line int, message string) {
	r.entries = append(r.entries, Entry{entryError, message, line})
}

// Warning adds a warning entry to the report.
func (r *Report) Warning(line int, message string) {
	r.entries = append(r.entries, Entry{entryWarning, message, line})
}

// Info adds an info entry to the report.
func (r *Report) Info(line int, message string) {
	r.entries = append(r.entries, Entry{entryInfo, message, line})
}

// Entries returns the list of entries in the report.
func (r *Report) Entries() []Entry {
	return r.entries
}

// Entry represents a single generic item in the report.
type Entry struct {
	kind    entryKind
	message string
	line    int
}

// String returns a human-readable representation of the entry.
func (e Entry) String() string {
	return fmt.Sprintf("line %d: %s: %s", e.line, e.kind, e.message)
}

// MarshalJSON satisfies the json.Marshaler interface, returning the entry
// encoded as a JSON object.
func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"kind":    e.kind.String(),
		"message": e.message,
		"line":    e.line,
	})
}

type entryKind int

const (
	entryError entryKind = iota
	entryWarning
	entryInfo
)

func (k entryKind) String() string {
	switch k {
	case entryError:
		return "error"
	case entryWarning:
		return "warning"
	case entryInfo:
		return "info"
	default:
		panic(fmt.Sprintf("invalid kind %d", k))
	}
}
