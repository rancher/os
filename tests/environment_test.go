package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestEnvironment(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_11/cloud-config.yml")

	s.CheckCall(c, "sudo system-docker inspect env | grep A=A")
	s.CheckCall(c, "sudo system-docker inspect env | grep BB=BB")
	s.CheckCall(c, "sudo system-docker inspect env | grep BC=BC")
}
