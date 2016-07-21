package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestSysctl(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_20/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo cat /proc/sys/kernel/domainname | grep test")
	s.CheckCall(c, "sudo cat /proc/sys/dev/cdrom/debug | grep 1")
}
