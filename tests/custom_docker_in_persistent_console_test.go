package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestCustomDockerInPersistentConsole(c *C) {
	err := s.RunQemu("--cloud-config", "./tests/assets/test_05/cloud-config.yml")
	c.Assert(err, IsNil)

	s.CheckCall(c, "curl", "-Lfo", "./docker", DockerUrl)
	s.CheckCall(c, "chmod", "+x", "/home/rancher/docker")
	s.CheckCall(c, "sudo", "ln", "-sf", "/home/rancher/docker", "/usr/bin/docker")
	s.CheckCall(c, "sudo", "system-docker", "restart", "docker")
	s.CheckCall(c, "sudo", "system-docker", "version")
}
