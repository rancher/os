package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestBadState(c *C) {
	err := s.RunQemu("--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE")
	c.Assert(err, IsNil)
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}

func (s *QemuSuite) TestBadStateWithWait(c *C) {
	err := s.RunQemu("--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE", "--append", "rancher.state.wait")
	c.Assert(err, IsNil)
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}
