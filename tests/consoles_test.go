package integration

import . "github.com/cpuguy83/check"

func (s *QemuSuite) TestCloudConfigConsole(c *C) {
	c.Parallel()
	err := s.RunQemu(c, "--cloud-config", "./tests/assets/test_03/cloud-config.yml")
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, "apt-get --version")
	s.CheckCall(c, `
sudo ros console list | grep default | grep disabled
sudo ros console list | grep debian | grep current`)
}

func (s *QemuSuite) TestConsoleCommand(c *C) {
	c.Parallel()
	err := s.RunQemu(c, )
	defer s.stopQemu(c)
	c.Assert(err, IsNil)

	s.CheckCall(c, `
sudo ros console list | grep default | grep current
sudo ros console list | grep debian | grep disabled`)

	s.MakeCall(c, "sudo ros console switch -f debian")
	c.Assert(s.WaitForSSH(c), IsNil)

	s.CheckCall(c, "apt-get --version")
	s.CheckCall(c, `
sudo ros console list | grep default | grep disabled
sudo ros console list | grep debian | grep current`)

	s.Reboot(c)

	s.CheckCall(c, "apt-get --version")
	s.CheckCall(c, `
sudo ros console list | grep default | grep disabled
sudo ros console list | grep debian | grep current`)

	s.MakeCall(c, "sudo ros console switch -f default")
	c.Assert(s.WaitForSSH(c), IsNil)

	s.CheckCall(c, `
sudo ros console list | grep default | grep current
sudo ros console list | grep debian | grep disabled`)

	s.CheckCall(c, "sudo ros console enable debian")

	s.CheckCall(c, "sudo ros console list | grep default | grep current")
	s.CheckCall(c, "sudo ros console list | grep debian | grep enabled")

	s.Reboot(c)

	s.CheckCall(c, `
sudo ros console list | grep default | grep disabled
sudo ros console list | grep debian | grep current`)
}
