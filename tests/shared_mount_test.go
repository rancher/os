package integration

import (
	"fmt"

	. "github.com/cpuguy83/check"
)

func (s *QemuSuite) TestSharedMount(c *C) {
	c.Parallel()
	err := s.RunQemu(c, )
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, fmt.Sprintf(`
set -x -e

sudo mkdir /mnt/shared
sudo touch /test
sudo system-docker run --privileged -v /mnt:/mnt:shared -v /test:/test %s mount --bind / /mnt/shared
ls /mnt/shared | grep test`, BusyboxImage))
}
