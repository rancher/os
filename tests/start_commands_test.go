package integration

import (
	"fmt"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestStartCommands(c *C) {
	s.RunQemu(c, "--cloud-config", "./tests/assets/test_26/cloud-config.yml")

	for i := 1; i < 6; i++ {
		s.CheckCall(c, fmt.Sprintf("ls /home/rancher | grep test%d", i))
	}
	s.CheckCall(c, "docker ps | grep nginx")
}
