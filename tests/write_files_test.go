package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestWriteFiles(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_24/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo cat /test | grep 'console content'")
	s.CheckCall(c, "sudo cat /test2 | grep 'console content'")
	s.CheckCall(c, "sudo system-docker exec ntp cat /test | grep 'ntp content'")
	s.CheckCall(c, "sudo system-docker exec syslog cat /test | grep 'syslog content'")
}
