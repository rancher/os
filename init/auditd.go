// +build linux

package init

import (
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
		if err := cmd.Start(); err != nil {
			log.Debug(err)
			return c, err
		}
	}
	return c, nil
}
