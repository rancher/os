package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestKernelHeaders(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_22/cloud-config.yml")

	s.CheckCall(c, `
sleep 30
sudo system-docker inspect kernel-headers`)
}
