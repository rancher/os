package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestWriteFiles(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_23/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo cat /test | grep 'console content'")
	s.CheckCall(c, "sudo cat /test2 | grep 'console content'")
	s.CheckCall(c, "sudo system-docker exec ntp cat /test | grep 'ntp content'")
	s.CheckCall(c, "sudo system-docker exec syslog cat /test | grep 'syslog content'")
}
