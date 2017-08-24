package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestCustomDocker(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_05/cloud-config.yml")

	s.CheckCall(c, `
set -ex

docker version | grep 1.12.6

sudo ros engine list | grep 1.12.6 | grep current
(sudo ros engine switch invalid 2>&1 || true) | grep "invalid is not a valid engine"
(sudo ros engine enable invalid 2>&1 || true) | grep "invalid is not a valid engine"

docker run -d --restart=always nginx
docker ps | grep nginx`)

	s.CheckCall(c, `
set -ex

sudo ros engine switch docker-1.13.1
/usr/sbin/wait-for-docker
docker version | grep 1.13.1
sudo ros engine list | grep 1.13.1 | grep current
docker ps | grep nginx`)

	s.Reboot(c)

	s.CheckCall(c, `
set -ex

docker version | grep 1.13.1
sudo ros engine list | grep 1.13.1 | grep current
docker ps | grep nginx`)
}

func (s *QemuSuite) TestCustomDockerInPersistentConsole(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_25/cloud-config.yml")

	s.CheckCall(c, `
set -ex

apt-get --version
docker version | grep 17.06.0-ce
sudo ros engine list | grep 17.06.0-ce | grep current
docker run -d --restart=always nginx
docker ps | grep nginx`)

	s.CheckCall(c, `
set -ex

sudo ros engine switch docker-1.12.6
/usr/sbin/wait-for-docker
docker version | grep 1.12.6
sudo ros engine list | grep 1.12.6 | grep current
docker ps | grep nginx`)

	s.Reboot(c)

	s.CheckCall(c, `
set -ex

docker version | grep 1.12.6
sudo ros engine list | grep 1.12.6 | grep current
docker ps | grep nginx`)
}
