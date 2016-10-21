package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestNetworkOnBoot(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_18/cloud-config.yml", "-net", "nic,vlan=1,model=virtio")
	c.Assert(err, IsNil)

	s.CheckCall(c, "apt-get --version")
	s.CheckCall(c, "sudo system-docker images | grep tianon/true")
}
