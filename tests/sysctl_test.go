package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestSysctl(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_20/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo cat /proc/sys/kernel/domainname | grep test")
	s.CheckCall(c, "sudo cat /proc/sys/dev/cdrom/debug | grep 1")
}
