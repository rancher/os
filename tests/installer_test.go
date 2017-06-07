package integration

import (
	"fmt"
	"strings"
	"time"

	. "gopkg.in/check.v1"
)

func (s *QemuSuite) TestInstallMsDosMbr(c *C) {
	// test_17 cloud config is an invalid http proxy cfg, so the installer has no network
	runArgs := []string{
		"--iso",
		"--fresh",
		"--cloud-config",
		"./tests/assets/test_17/cloud-config.yml",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)
		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
echo "---------------------------------- generic"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append rancher.password=rancher
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.Stop(c)
}

func (s *QemuSuite) TestInstallGptMbr(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
echo "---------------------------------- gptsyslinux"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -t gptsyslinux -c config.yml
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	// TEST parted output? (gpt non-uefi == legacy_boot)
	s.Stop(c)
}

func (s *QemuSuite) TestInstallAlpine(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	{
		s.RunQemuWith(c, runArgs...)

		s.MakeCall("sudo ros console switch -f alpine")
		c.Assert(s.WaitForSSH(), IsNil)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
set -ex
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.Stop(c)
}

func (s *QemuSuite) TestAutoResize(c *C) {
	runArgs := []string{
		"--iso",
		"--fresh",
	}
	version := ""
	disk := "/dev/vda1\n"
	size := ""
	{
		s.RunQemuWith(c, runArgs...)

		version = s.CheckOutput(c, version, Not(Equals), "sudo ros -v")
		fmt.Printf("installing %s", version)

		s.CheckCall(c, `
set -ex
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append "rancher.resize_device=/dev/vda"
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.CheckCall(c, "sudo mount "+strings.TrimSpace(disk)+" /mnt")
		size = s.CheckOutput(c, size, Not(Equals), "df -h | grep "+strings.TrimSpace(disk)+" | head -n1 | sed 's/ \\+/;/g' | cut -d ';' -f 2")
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
		"--resizehd", "+20G",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.CheckOutput(c, disk, Equals, "sudo blkid | cut -f 1 -d ' ' | sed 's/://'")
	s.CheckOutput(c, size, Not(Equals), "df -h | grep "+strings.TrimSpace(disk)+" | head -n1 | sed 's/ \\+/;/g' | cut -d ';' -f 2")

	s.Stop(c)
}
