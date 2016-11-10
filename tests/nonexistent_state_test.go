package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestNonexistentState(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--no-format", "--append", "rancher.state.dev=LABEL=NONEXISTENT")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo ros config get rancher.state.dev | grep LABEL=NONEXISTENT")
}
