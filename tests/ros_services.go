package integration

import check "gopkg.in/check.v1"

func (s *QemuSuite) TestRosLocalService(c *check.C) {
	s.RunQemu(c)

	// System-docker
	s.CheckCall(c, `echo "FROM $(sudo system-docker images --format '{{.Repository}}:{{.Tag}}' | grep os-base)" > Dockerfile
sudo system-docker build -t testimage .`)

	s.CheckCall(c, `echo "test:" > test.yml
echo "  image: testimage" >> test.yml
echo "  entrypoint: ls" >> test.yml
echo "  labels:" >> test.yml
echo "    io.rancher.os.scope: system" >> test.yml
echo "    io.rancher.os.after: console" >> test.yml
`)

	s.CheckCall(c, `sudo cp test.yml /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service enable /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service up test`)

	s.CheckCall(c, `sudo ros service logs test | grep bin`)
}

func (s *QemuSuite) TestRosLocalServiceUser(c *check.C) {
	s.RunQemu(c)

	// User-docker
	s.CheckCall(c, `echo "FROM alpine" > Dockerfile
sudo docker build -t testimage .`)

	s.CheckCall(c, `echo "test:" > test.yml
echo "  image: testimage" >> test.yml
echo "  entrypoint: ls" >> test.yml
echo "  labels:" >> test.yml
echo "    io.rancher.os.scope: user" >> test.yml
echo "    io.rancher.os.after: console" >> test.yml
`)

	s.CheckCall(c, `sudo cp test.yml /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service enable /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service up test`)

	s.CheckCall(c, `sudo ros service logs test | grep bin`)
}
