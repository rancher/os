package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestSwap(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_21/cloud-config.yml", "--second-drive")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo mkswap /dev/vdb")
	s.CheckCall(c, "sudo cloud-init-execute")
	s.CheckCall(c, "cat /proc/swaps | grep /dev/vdb")
}
