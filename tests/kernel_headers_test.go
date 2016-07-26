package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestKernelHeaders(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_22/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
sleep 10
docker inspect kernel-headers`)
}
