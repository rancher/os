package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestTls(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_02/cloud-config.yml")
	s.CheckCall(c, `
set -e -x
sudo ros tls gen
docker --tlsverify version`)
}
