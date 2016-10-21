package integration

import (
	"fmt"

	. "github.com/cpuguy83/check"
)

func (s *QemuSuite) TestUpgrade(c *C) {
	c.Parallel()
	//TODO: this will fail if its the first time we've ever run?
	err := s.RunQemuInstalled(c, )
	c.Assert(err, IsNil)

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
