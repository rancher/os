package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestEnvironment(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_11/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo system-docker inspect env | grep A=A")
	s.CheckCall(c, "sudo system-docker inspect env | grep BB=BB")
	s.CheckCall(c, "sudo system-docker inspect env | grep BC=BC")
}
