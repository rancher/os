package integration

import (
	"fmt"
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

func (s *QemuSuite) TestInstallCustomConsoleWithLowMemory(c *C) {
	// this test shows that when the installed HD boots, it quietly fails to use the user_config
	// we basically boot ignoring all the user's requests

	// test_17 cloud config is an invalid http proxy cfg, so the installer has no network
	runArgs := []string{
		"--memory",
		"768",
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

// TODO: test with `# cloud-config` - that failed with 0.8.0
		s.CheckCall(c, `
echo "---------------------------------- generic"
set -ex
sudo parted /dev/vda print
touch config.yml
echo "#cloud-config" >> config.yml
echo "rancher:" >> config.yml
echo "  console: ubuntu" >> config.yml
echo "ssh_authorized_keys:" >> config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml --append rancher.password=rancher
sync
`)
		time.Sleep(500 * time.Millisecond)
		s.Stop(c)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--memory",
		"512",
		"--boothd",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, version, Equals, "sudo ros -v")
	s.CheckCall(c, "which apt-get")
	s.CheckCall(c, "free")
	s.CheckCall(c, "sudo system-docker ps")


	s.Stop(c)
}
