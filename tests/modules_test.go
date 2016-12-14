package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestKernelParameterModule(c *C) {
	s.RunQemu(c, "--append", "rancher.modules=[btrfs]")
	s.CheckCall(c, "lsmod | grep btrfs")
}

func (s *QemuSuite) TestCloudConfigModule(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_27/cloud-config.yml")
	s.CheckCall(c, "lsmod | grep btrfs")
}
