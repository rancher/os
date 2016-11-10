package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestCloudConfigHostname(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_13/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "hostname | grep rancher-test")
	s.CheckCall(c, "cat /etc/hosts | grep rancher-test")
}
