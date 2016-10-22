package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestHttpProxy(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_17/cloud-config.yml")

	s.CheckCall(c, `
set -x -e

sudo system-docker inspect docker | grep HTTP_PROXY=invalid
sudo system-docker inspect docker | grep HTTPS_PROXY=invalid
sudo system-docker inspect docker | grep NO_PROXY=invalid

if docker pull busybox; then
    exit 1
else
    exit 0
fi`)

	s.Reboot(c)

	s.CheckCall(c, `
set -x -e

if sudo system-docker pull busybox; then
    exit 1
else
    exit 0
fi`)
}
