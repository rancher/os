package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestSubdir(c *C) {
	s.RunQemu(c, "--append", "rancher.state.directory=ros_subdir")
	s.CheckCall(c, `
set -x -e
mkdir x
sudo mount $(sudo ros dev LABEL=RANCHER_STATE) x
[ -d x/ros_subdir/home/rancher ]`)
}
