// +build linux

package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func initializeAuditd(c *config.CloudConfig) (*config.CloudConfig, error) {
	for _, d := range []string{"/var/log/audit"} {
		if err := os.MkdirAll(d, os.ModeDir); err != nil {
			log.Debug(err)
			return c, err
		}
	}
	cmds := []*exec.Cmd{
		exec.Command("auditctl", "-R", "/etc/audit/audit.rules"),
		exec.Command("auditd", "-n"),
	}
	for _, cmd := range cmds {
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Debug(err)
			return c, err
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Debug(err)
			return c, err
		}

		if err := cmd.Start(); err != nil {
			log.Debug(err)
			return c, err
		}

		outs, _ := ioutil.ReadAll(stdout)
		errs, _ := ioutil.ReadAll(stderr)

		if err := cmd.Wait(); err != nil {
			fmt.Printf("%s\n", outs)
			fmt.Printf("%s\n", errs)
			log.Fatalf("%s : %s", cmd.Path, err)
		}
	}
	return c, nil
}
