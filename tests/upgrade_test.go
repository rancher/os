package integration

import (
	"fmt"
	"time"

	"strings"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestUpgrade050(c *C) {
	s.commonTestCode(c, "v0.5.0", "debian")
}
func (s *QemuSuite) TestUpgrade070(c *C) {
	s.commonTestCode(c, "v0.7.0", "default")
}
func (s *QemuSuite) TestUpgrade071(c *C) {
	s.commonTestCode(c, "v0.7.1", "ubuntu")
}
func (s *QemuSuite) TestUpgrade080rc1(c *C) {
	s.commonTestCode(c, "v0.8.0-rc1", "alpine")
}
func (s *QemuSuite) TestUpgrade080rc7(c *C) {
	s.commonTestCode(c, "v0.8.0-rc7", "centos")
}

func (s *QemuSuite) commonTestCode(c *C, startWithVersion, console string) {
	runArgs := []string{
		"--iso",
		"--fresh",
		"--cloud-config",
		"./tests/assets/test_12/cloud-config.yml",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)
		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		version = strings.TrimSpace(strings.TrimPrefix(version, "ros version"))
		c.Assert(Version, Equals, version)

		fmt.Printf("installing %s", startWithVersion)
		s.PullAndLoadInstallerImage(c, fmt.Sprintf("rancher/os:%s", startWithVersion))

		//ADD a custom append line and make sure its kept in the upgraded version too

		s.CheckCall(c, fmt.Sprintf(`
echo "---------------------------------- generic"
set -ex
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append rancher.password=rancher -i rancher/os:%s
sync
		`, startWithVersion))
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, "ros version "+startWithVersion+"\n", Equals, "sudo ros -v")

	if console != "default" {
		// Can't preload the startWithVersion console image, as some don't exist by that name - not sure how to approach that
		//s.PullAndLoadInstallerImage(c, fmt.Sprintf("rancher/os-%sconsole:%s", console, startWithVersion))
		// TODO: ouch. probably need to tag the dev / master version as latest cos this won't work
		// Need to pull the image here - if we do it at boot, then the test will fail.
		s.PullAndLoadInstallerImage(c, fmt.Sprintf("rancher/os-%sconsole:%s", console, "v0.8.0-rc3"))
		s.CheckCall(c, fmt.Sprintf("sudo ros console enable %s", console))
	}

	s.LoadInstallerImage(c)
	s.CheckCall(c, fmt.Sprintf("sudo ros os upgrade --no-reboot -i rancher/os:%s%s --force", Version, Suffix))

	s.Reboot(c)
	s.CheckOutput(c, "ros version "+Version+"\n", Equals, "sudo ros -v")

	s.Stop(c)
}
