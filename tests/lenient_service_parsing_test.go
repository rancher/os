package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestLenientServiceParsing(c *C) {
	c.Parallel()
	err := s.RunQemu("--cloud-config", "./tests/assets/test_19/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
sleep 5
sudo system-docker ps -a | grep test-parsing`)
}
