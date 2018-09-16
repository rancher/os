// +build linux

package util

import "github.com/tredoe/term"

func IsRunningInTty() bool {
	return term.IsTerminal(1)
}
