package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestRebootWithContainerRunning(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_03/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, fmt.Sprintf(`
set -e -x
docker run -d --restart=always %s`, NginxImage))

	s.Reboot()

	s.CheckCall(c, "docker ps -f status=running | grep nginx")
}
