package integration

import (
	. "github.com/cpuguy83/check"
)

func (s *QemuSuite) TestPreload(c *C) {
	c.Parallel()
	err := s.RunQemu(c, )
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, `
docker pull busybox
sudo docker save -o /var/lib/rancher/preload/system-docker/busybox.tar busybox
sudo gzip /var/lib/rancher/preload/system-docker/busybox.tar
sudo system-docker pull alpine
sudo system-docker save -o /var/lib/rancher/preload/docker/alpine.tar alpine`)

	s.Reboot(c)

	s.CheckCall(c, `
sleep 5
sudo system-docker images | grep busybox
docker images | grep alpine`)
}
