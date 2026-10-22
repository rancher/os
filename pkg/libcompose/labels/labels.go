package labels

import (
	"fmt"

	"github.com/burmilla/os/pkg/libcompose/utils"
)

// Label represents a docker label.
type Label string

// Libcompose default labels.
const (
	NUMBER  = Label("com.docker.compose.container-number")
	ONEOFF  = Label("com.docker.compose.oneoff")
	PROJECT = Label("com.docker.compose.project")
	SERVICE = Label("com.docker.compose.service")
	HASH    = Label("com.docker.compose.config-hash")
	VERSION = Label("com.docker.compose.version")

	PROJECT_LEGACY = Label("io.docker.compose.project")
	SERVICE_LEGACY = Label("io.docker.compose.service")
	HASH_LEGACY    = Label("io.docker.compose.config-hash")
)

// EqString returns a label json string representation with the specified value.
func (f Label) EqString(value string) string {
	return LabelFilterString(string(f), value)
}

// Eq returns a label map representation with the specified value.
func (f Label) Eq(value string) map[string][]string {
	return LabelFilter(string(f), value)
}

// Str returns the label name.
func (f Label) Str() string {
	return string(f)
}

// LabelFilterString returns a label json string representation of the specifed couple (key,value)
// that is used as filter for docker.
func LabelFilterString(key, value string) string {
	return utils.FilterString(map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	})
}

// LabelFilter returns a label map representation of the specifed couple (key,value)
// that is used as filter for docker.
func LabelFilter(key, value string) map[string][]string {
	return map[string][]string{
		"label": {fmt.Sprintf("%s=%s", key, value)},
	}
}
