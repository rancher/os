package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestLenientServiceParsing(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_19/cloud-config.yml")

	s.CheckCall(c, `
sleep 5
sudo system-docker ps -a | grep test-parsing`)
}
