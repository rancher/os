package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestWriteFiles(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_24/cloud-config.yml")

	s.CheckCall(c, "sudo cat /test | grep 'console content'")
	s.CheckCall(c, "sudo cat /test2 | grep 'console content'")
	s.CheckCall(c, "sudo system-docker exec ntp cat /test | grep 'ntp content'")
	s.CheckCall(c, "sudo system-docker exec syslog cat /test | grep 'syslog content'")
}
