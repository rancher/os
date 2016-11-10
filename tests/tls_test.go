package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestTls(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_02/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, `
set -e -x
sudo ros tls gen
docker --tlsverify version`)
}
