package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestSharedMount(c *C) {
	s.RunQemu(c)
	s.CheckCall(c, fmt.Sprintf(`
set -x -e

sudo mkdir /mnt/shared
sudo touch /test
sudo system-docker run --privileged -v /mnt:/mnt:shared -v /test:/test %s mount --bind / /mnt/shared
ls /mnt/shared | grep test`, BusyboxImage))
}
