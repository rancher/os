package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestBadState(c *C) {
	s.RunQemu(c, "--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE")
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}

func (s *QemuSuite) TestBadStateWithWait(c *C) {
	s.RunQemu(c, "--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE", "--append", "rancher.state.wait")
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}
