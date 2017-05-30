// +build linux

package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

const rulesFile = "/etc/audit/audit.rules"

func initializeAuditd(c *config.CloudConfig) (*config.CloudConfig, error) {
	if c.Rancher.Auditd.Enabled {
		for _, d := range []string{"/var/log/audit", path.Dir(rulesFile)} {
			if err := os.MkdirAll(d, os.ModeDir); err != nil {
				log.Debug(err)
				return c, err
			}
		}
		if err := writeRules(&c.Rancher.Auditd); err != nil {
			return c, fmt.Errorf("Error writing audit rules: %s", err)
		}
		cmds := []*exec.Cmd{
			exec.Command("auditctl", "-R", rulesFile),
			exec.Command("auditd", "-n"),
		}
		for _, cmd := range cmds {
			if err := cmd.Start(); err != nil {
				log.Debug(err)
				return c, err
			}
		}
	}
	return c, nil
}

func writeRules(r *config.AuditConfig) error {
	rules := strings.Join(r.Rules, "\n")

	if err := ioutil.WriteFile(rulesFile, []byte(rules), 0400); err != nil {
		return err
	}
	return nil
}
