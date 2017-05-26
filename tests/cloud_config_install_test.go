package integration

import (
	"time"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestCloudConfigInstall(c *C) {
	s.RunQemu(c,
		"--iso",
		"--fresh",
		"--no-format",
		"--cloud-config", "./tests/assets/cloud_config_install_test/cloud-config.yml")

	//check we have a particular version, from iso
	s.CheckOutput(c, " Backing Filesystem: tmpfs\n", Equals, "sudo system-docker info | grep Filesystem")
	//and no persistence yet
	//s.CheckOutput(c, "\n", Equals, "sudo blkid")
	// TODO: need some way to wait for install to complete.
	time.Sleep(time.Second)
	for {
		result, _ := s.MakeCall("cat", "/var/log/ros-install.log")
		if result == "done\n" {
			break
		}
		time.Sleep(time.Second * 3)
	}
	//check we have persistence and that ros-install completed ok
	s.CheckOutput(c, "/dev/vda1:\n", Equals, "sudo blkid | grep RANCHER_STATE | cut -d ' ' -f 1")
	s.CheckOutput(c, "LABEL=\"RANCHER_STATE\"\n", Equals, "sudo blkid | grep vda1 | cut -d ' ' -f 2")

	//reboot, and check we're using the new non-iso install
	s.Stop(c)
	s.RunQemuWith(c, "--qemu", "--boothd", "--no-rm-usr")
	s.CheckOutput(c, "/dev/vda1:\n", Equals, "sudo blkid | grep RANCHER_STATE | cut -d ' ' -f 1")
	s.CheckOutput(c, "LABEL=\"RANCHER_STATE\"\n", Equals, "sudo blkid | grep vda1 | cut -d ' ' -f 2")
	s.CheckOutput(c, " Backing Filesystem: extfs\n", Equals, "sudo system-docker info | grep Filesystem")
}
