package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestNetworkOnBoot(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_18/cloud-config.yml", "-net", "nic,vlan=1,model=virtio")

	s.CheckCall(c, "apt-get --version")
	s.CheckCall(c, "sudo system-docker images | grep tianon/true")
}
