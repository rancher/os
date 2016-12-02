package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestNonexistentState(c *C) {
	s.RunQemu(c, "--no-format", "--append", "rancher.state.dev=LABEL=NONEXISTENT")
	s.CheckCall(c, "sudo ros config get rancher.state.dev | grep LABEL=NONEXISTENT")
}
