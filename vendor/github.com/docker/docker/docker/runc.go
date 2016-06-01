// +build !exclude_runc

package docker

import (
	"github.com/docker/docker/pkg/reexec"
	"github.com/opencontainers/runc"
)

func init() {
	reexec.Register("docker-runc", runc.Main)
}
