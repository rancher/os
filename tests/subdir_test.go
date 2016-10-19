package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestSubdir(c *C) {
	c.Parallel()
	err := s.RunQemu("--append", "rancher.state.directory=ros_subdir")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -x -e
mkdir x
sudo mount $(sudo ros dev LABEL=RANCHER_STATE) x
[ -d x/ros_subdir/home/rancher ]`)
}
