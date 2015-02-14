package util

import (
	"github.com/kless/term"

	log "github.com/Sirupsen/logrus"
)

func IsRunningInTty() bool {
	log.Infof("Is a tty : %v", term.IsTerminal(0))
	log.Infof("Is a tty : %v", term.IsTerminal(1))
	log.Infof("Is a tty : %v", term.IsTerminal(2))
	return term.IsTerminal(1)
}
