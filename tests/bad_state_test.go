package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestBadState(c *C) {
	c.Parallel()
	err := s.RunQemu("--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE")
	c.Assert(err, IsNil)
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}

func (s *QemuSuite) TestBadStateWithWait(c *C) {
	c.Parallel()
	err := s.RunQemu("--no-format", "--append", "rancher.state.dev=LABEL=BAD_STATE", "--append", "rancher.state.wait")
	c.Assert(err, IsNil)
	s.CheckCall(c, "mount | grep /var/lib/docker | grep rootfs")
}
