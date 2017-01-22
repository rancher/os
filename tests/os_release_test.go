package integration

import (
	. "gopkg.in/check.v1"
	"strings"
)

func (s *QemuSuite) TestOsRelease(c *C) {
	s.RunQemu(c)
	version := ""
	version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
	version = strings.Replace(version, "ros version ", "", 1)
	s.CheckOutput(c, "VERSION="+version, Equals, "cat /etc/os-release | grep VERSION=")
	s.CheckOutput(c, "NAME=\"RancherOS\"\n", Equals, "cat /etc/os-release | grep ^NAME=")

	s.MakeCall("sudo ros console switch -f alpine")
	c.Assert(s.WaitForSSH(), IsNil)

	s.CheckOutput(c, "/sbin/apk\n", Equals, "which apk")
	s.CheckOutput(c, "VERSION="+version, Equals, "cat /etc/os-release | grep VERSION=")
	s.CheckOutput(c, "NAME=\"RancherOS\"\n", Equals, "cat /etc/os-release | grep ^NAME=")

	s.Reboot(c)

	s.CheckOutput(c, "/sbin/apk\n", Equals, "which apk")
	s.CheckOutput(c, "VERSION="+version, Equals, "cat /etc/os-release | grep VERSION=")
	s.CheckOutput(c, "NAME=\"RancherOS\"\n", Equals, "cat /etc/os-release | grep ^NAME=")
}
