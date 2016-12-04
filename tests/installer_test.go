package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestInstall(c *C) {
	// ./scripts/run --no-format --append "rancher.debug=true"  --iso --fresh
	runArgs := []string{
		"--iso",
		"--fresh",
		"--no-format",
		"--append", "rancher.debug=true",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckCall(c, `
set -ex
sudo ros install --force --no-reboot --device /dev/vda`)

	s.Stop(c)

	// ./scripts/run --no-format --append "rancher.debug=true"
	runArgs = []string{
		"--no-format",
		"--append", "rancher.debug=true",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckCall(c, "sudo ros -v")
}
