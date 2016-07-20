package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestCloudConfigMounts(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_16/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, "cat /home/rancher/test | grep test")
}
