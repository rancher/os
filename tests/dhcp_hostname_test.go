package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestDhcpHostname(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_12/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "hostname | grep rancher-dev")
	s.CheckCall(c, "cat /etc/hosts | grep rancher-dev")
}
