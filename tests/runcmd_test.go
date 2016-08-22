package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestRuncmd(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_26/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, "ls /home/rancher | grep test")
	s.CheckCall(c, "ls /home/rancher | grep test2")
}
