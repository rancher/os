package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestUpgrade(c *C) {
	s.RunQemuInstalled(c)

	s.CheckCall(c, `
set -ex
sudo ros os upgrade -i rancher/os:v0.5.0 --force --no-reboot`)

	s.Reboot(c)

	s.CheckCall(c, "sudo ros -v | grep v0.5.0")
	s.LoadInstallerImage(c)
	s.CheckCall(c, fmt.Sprintf("sudo ros os upgrade -i rancher/os:%s%s --force --no-reboot", Version, Suffix))

	s.Reboot(c)

	s.CheckCall(c, fmt.Sprintf("sudo ros -v | grep %s", Version))
}
