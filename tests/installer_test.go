package integration

import . "gopkg.in/check.v1"

// TODO: separate out into different tests - there's something that makes one pass and one fail.

//func (s *QemuSuite) TestInstallMsDosMbr(c *C) {
func (s *QemuSuite) TestInstall(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
		"--nodisplay",
	}
	{
		s.RunQemuWith(c, runArgs...)
		defer s.Stop(c)

		s.CheckCall(c, `
echo "---------------------------------- generic"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -c config.yml`)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
		"--nodisplay",
	}
	s.RunQemuWith(c, runArgs...)
	defer s.Stop(c)

	s.CheckCall(c, "sudo ros -v")
	//}

	//func (s *QemuSuite) TestInstallGptMbr(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs = []string{
		"--iso",
		"--fresh",
		"--nodisplay",
	}
	{
		s.RunQemuWith(c, runArgs...)
		defer s.Stop(c)

		s.CheckCall(c, `
echo "---------------------------------- gptsyslinux"
set -ex
sudo parted /dev/vda print
echo "ssh_authorized_keys:" > config.yml
echo "  - $(cat /home/rancher/.ssh/authorized_keys)" >> config.yml
sudo ros install --force --no-reboot --device /dev/vda -t gptsyslinux -c config.yml`)
	}

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--boothd",
		"--nodisplay",
	}
	s.RunQemuWith(c, runArgs...)
	defer s.Stop(c)

	s.CheckCall(c, "sudo ros -v")
	// TEST parted output? (gpt non-uefi == legacy_boot)
}
