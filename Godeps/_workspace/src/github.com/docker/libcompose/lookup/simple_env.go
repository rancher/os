package lookup

import (
	"fmt"
	"os"

	"github.com/docker/libcompose/project"
)

type OsEnvLookup struct {
}

func (o *OsEnvLookup) Lookup(key, serviceName string, config *project.ServiceConfig) []string {
	ret := os.Getenv(key)
	if ret == "" {
		return []string{}
	} else {
		return []string{fmt.Sprintf("%s=%s", key, ret)}
	}
}
