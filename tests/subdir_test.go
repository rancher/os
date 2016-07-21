package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestSubdir(c *C) {
	err := s.RunQemu("--append", "rancher.state.directory=ros_subdir")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -x -e
mkdir x
sudo mount $(sudo ros dev LABEL=RANCHER_STATE) x
[ -d x/ros_subdir/home/rancher ]`)
}
