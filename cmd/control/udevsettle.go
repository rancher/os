package control

import (
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func udevSettleAction(c *cli.Context) {
	if err := UdevSettle(); err != nil {
		log.Fatal(err)
	}
}

func UdevSettle() error {
	cmd := exec.Command("udevd", "--daemon")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("udevadm", "trigger", "--action=add")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("udevadm", "settle")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
