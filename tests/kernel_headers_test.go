package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestKernelHeaders(c *C) {
	c.Parallel()
	err := s.RunQemu("--cloud-config", "./tests/assets/test_22/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
sleep 15
docker inspect kernel-headers`)
}
