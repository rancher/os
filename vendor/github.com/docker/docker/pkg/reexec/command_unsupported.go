// +build !linux,!windows,!freebsd

package reexec

import (
	"github.com/docker/containerd/subreaper/exec"
)

// Command is unsupported on operating systems apart from Linux and Windows.
func Command(args ...string) *exec.Cmd {
	return nil
}
