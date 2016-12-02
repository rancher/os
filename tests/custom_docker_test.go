package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestCustomDocker(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_05/cloud-config.yml")

	s.CheckCall(c, `
set -ex

docker version | grep 1.10.3

sudo ros engine list | grep 1.10.3 | grep current
docker run -d --restart=always nginx
docker ps | grep nginx`)

	s.CheckCall(c, `
set -ex

sudo ros engine switch docker-1.11.2
/usr/sbin/wait-for-docker
docker version | grep 1.11.2
sudo ros engine list | grep 1.11.2 | grep current
docker ps | grep nginx`)

	s.Reboot(c)

	s.CheckCall(c, `
set -ex

docker version | grep 1.11.2
sudo ros engine list | grep 1.11.2 | grep current
docker ps | grep nginx`)
}

func (s *QemuSuite) TestCustomDockerInPersistentConsole(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_25/cloud-config.yml")

	s.CheckCall(c, `
set -ex

apt-get --version
docker version | grep 1.10.3
sudo ros engine list | grep 1.10.3 | grep current
docker run -d --restart=always nginx
docker ps | grep nginx`)

	s.CheckCall(c, `
set -ex

sudo ros engine switch docker-1.11.2
/usr/sbin/wait-for-docker
docker version | grep 1.11.2
sudo ros engine list | grep 1.11.2 | grep current
docker ps | grep nginx`)

	s.Reboot(c)

	s.CheckCall(c, `
set -ex

docker version | grep 1.11.2
sudo ros engine list | grep 1.11.2 | grep current
docker ps | grep nginx`)
}
