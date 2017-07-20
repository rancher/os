package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestRosLocalService(c *C) {
	s.RunQemu(c)

	s.CheckCall(c, `echo "FROM $(sudo system-docker images --format '{{.Repository}}:{{.Tag}}' | grep os-base)" > Dockerfile
sudo system-docker build -t testimage .

echo "test:" > test.yml
echo "  image: testimage" >> test.yml
echo "  entrypoint: ls" >> test.yml
echo "  labels:" >> test.yml
echo "    io.rancher.os.scope: system" >> test.yml
echo "    io.rancher.os.after: console" >> test.yml
`)

	s.CheckCall(c, `sudo cp test.yml /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service enable /var/lib/rancher/conf/test.yml`)
	s.CheckCall(c, `sudo ros service up /var/lib/rancher/conf/test.yml`)

	s.CheckCall(c, `sudo ros service logs test | grep bin`)
}
