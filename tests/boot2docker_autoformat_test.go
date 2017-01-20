package integration

import . "gopkg.in/check.v1"

// NOTE: the boot2docker magic string/tar file is also used by Docker-machine

func (s *QemuSuite) TestBoot2DockerState(c *C) {
	s.RunQemu(c, "--fresh", "--b2d")
	s.CheckCall(c, "blkid | grep B2D_STATE")
	// And once I make run create a tar file, check that its untarred in the docker user's home dir
	// And confirm if it should add to the dir, or replace, i can't remember
}

func (s *QemuSuite) TestIsoBoot2DockerState(c *C) {
	s.RunQemu(c, "--fresh", "--b2d", "--iso")
	s.CheckCall(c, "blkid | grep B2D_STATE")
	s.CheckCall(c, "sudo ros config get rancher.state.dev | grep LABEL=B2D_STATE")
	s.CheckCall(c, "sudo ros config get rancher.state.autoformat | grep vda")
	// And once I make run create a tar file, check that its untarred in the docker user's home dir
	// And confirm if it should add to the dir, or replace, i can't remember
}

func (s *QemuSuite) TestRancherOSState(c *C) {
	s.RunQemu(c, "--fresh")
	s.CheckCall(c, "blkid | grep RANCHER_STATE")
}
