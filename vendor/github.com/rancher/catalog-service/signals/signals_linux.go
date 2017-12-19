// +build linux

package signals

import (
	"syscall"

	log "github.com/Sirupsen/logrus"
)

func init() {
	if _, _, err := syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(syscall.SIGTERM), 0); err != 0 {
		log.Fatalf("Failed to set parent death signal: %d", err)
	}
}
