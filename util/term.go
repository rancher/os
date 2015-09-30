// +build linux

package util

import "github.com/kless/term"

func IsRunningInTty() bool {
	return term.IsTerminal(1)
}
