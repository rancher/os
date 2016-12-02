package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestSwap(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_21/cloud-config.yml", "--second-drive")

	s.CheckCall(c, "sudo mkswap /dev/vdb")
	s.CheckCall(c, "sudo cloud-init-execute")
	s.CheckCall(c, "cat /proc/swaps | grep /dev/vdb")
}
