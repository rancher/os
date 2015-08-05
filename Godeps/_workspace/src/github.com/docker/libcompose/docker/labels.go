package docker

import (
	"encoding/json"

	"github.com/docker/libcompose/utils"
)

type Label string

const (
	NAME    = Label("io.docker.compose.name")
	PROJECT = Label("io.docker.compose.project")
	SERVICE = Label("io.docker.compose.service")
	HASH    = Label("io.docker.compose.config-hash")
	REBUILD = Label("io.docker.compose.rebuild")
)

func (f Label) Eq(value string) string {
	return utils.LabelFilter(string(f), value)
}

func And(left, right string) string {
	leftMap := map[string][]string{}
	rightMap := map[string][]string{}

	// Ignore errors
	json.Unmarshal([]byte(left), &leftMap)
	json.Unmarshal([]byte(right), &rightMap)

	for k, v := range rightMap {
		existing, ok := leftMap[k]
		if ok {
			leftMap[k] = append(existing, v...)
		} else {
			leftMap[k] = v
		}
	}

	result, _ := json.Marshal(leftMap)

	return string(result)
}

func (f Label) Str() string {
	return string(f)
}
