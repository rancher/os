package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestTls(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_02/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -e -x
sudo ros tls gen
docker --tlsverify version`)
}
