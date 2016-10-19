package integration

import (
	"fmt"

	. "github.com/cpuguy83/check"
)

func (s *QemuSuite) TestStartCommands(c *C) {
	c.Parallel()
	err := s.RunQemu("--cloud-config", "./tests/assets/test_26/cloud-config.yml")
	c.Assert(err, IsNil)

	for i := 1; i < 5; i++ {
		s.CheckCall(c, fmt.Sprintf("ls /home/rancher | grep test%d", i))
	}
	s.CheckCall(c, "docker ps | grep nginx")
}
