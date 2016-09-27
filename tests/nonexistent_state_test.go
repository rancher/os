package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestNonexistentState(c *C) {
	err := s.RunQemu("--no-format", "--append", "rancher.state.dev=LABEL=NONEXISTENT")
	c.Assert(err, IsNil)

	s.CheckCall(c, "sudo ros config get rancher.state.dev | grep LABEL=NONEXISTENT")
}
