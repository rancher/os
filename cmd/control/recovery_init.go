package control

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/burmilla/os/pkg/log"

	"github.com/codegangsta/cli"
)

func recoveryInitAction(c *cli.Context) error {
	if err := writeRespawn("root", false, true); err != nil {
		log.Error(err)
	}

	respawnBinPath, err := exec.LookPath("respawn")
	if err != nil {
		return err
	}

	return syscall.Exec(respawnBinPath, []string{"respawn", "-f", "/etc/respawn.conf"}, os.Environ())
}
