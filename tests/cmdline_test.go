package integration

import . "gopkg.in/check.v1"

func (s *QemuSuite) TestElideCmdLine(c *C) {
	runArgs := []string{
		"--fresh",
		"--append-init",
		"cc.hostname=nope rancher.debug=true",
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, "nope\n", Equals, "hostname")
	s.CheckOutput(c, "printk.devkmsg=on rancher.debug=true rancher.password=rancher console=ttyS0 rancher.autologin=ttyS0  rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda,/dev/vda] rancher.rm_usr -- \n", Equals, "cat /proc/cmdline")
}
