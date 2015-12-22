// +build linux

package init

import (
	"os"
	"os/signal"
	"syscall"
)

func pidOne() error {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGCHLD)

	var (
		ws  syscall.WaitStatus
		rus syscall.Rusage
	)
	for range c {
		for {
			if pid, err := syscall.Wait4(-1, &ws, syscall.WNOHANG, &rus); err != nil || pid <= 0 {
				break
			}
		}
	}

	return nil
}
