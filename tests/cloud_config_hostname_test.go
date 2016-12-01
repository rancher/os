package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestCloudConfigHostname(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_13/cloud-config.yml")

	s.CheckCall(c, "hostname | grep rancher-test")
	s.CheckCall(c, "cat /etc/hosts | grep rancher-test")
}
