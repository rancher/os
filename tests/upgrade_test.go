package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

// DisabledTestUpgrade, The new go based installer code breaks downgrading from itself to a previous version
// because 0.8.0 now uses syslinx and a set of syslinux.cfg files, whereas before that , we used grub and
// assumed that there was only one kernel&initrd
//      see installer_test.go for more tests
func (s *QemuSuite) DisabledTestUpgrade(c *C) {
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
