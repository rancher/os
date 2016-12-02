package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestMounts(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_23/cloud-config.yml", "--second-drive")

	s.CheckCall(c, "cat /home/rancher/test | grep test")

	s.CheckCall(c, "mkdir -p /home/rancher/a /home/rancher/b /home/rancher/c")
	s.CheckCall(c, "sudo mkfs.ext4 /dev/vdb")
	s.CheckCall(c, "sudo cloud-init-execute")
	s.CheckCall(c, "mount | grep /home/rancher/a")
	s.CheckCall(c, "mount | grep /home/rancher/b")
	s.CheckCall(c, "mount | grep /home/rancher/c")
}
