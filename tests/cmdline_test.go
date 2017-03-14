package integration

import . "gopkg.in/check.v1"
import "fmt"

func (s *QemuSuite) TestElideCmdLine(c *C) {
	extra := "cc.hostname=nope rancher.password=three"
	runArgs := []string{
		"--fresh",
		"--append",
		"cc.something=yes rancher.password=two",
		"--append-init",
		extra,
	}
	s.RunQemuWith(c, runArgs...)

	s.CheckOutput(c, "nope\n", Equals, "hostname")
	s.CheckOutput(c,
		"printk.devkmsg=on rancher.debug=true rancher.password=rancher console=ttyS0 rancher.autologin=ttyS0  cc.something=yes rancher.password=two rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda,/dev/vda] rancher.rm_usr -- \n",
		Equals,
		"cat /proc/cmdline",
	)
	s.CheckOutput(c,
		fmt.Sprintf("/init %s\n", extra),
		Equals,
		"sudo ros config get rancher.environment.EXTRA_CMDLINE",
	)
	// TODO: it seems that rancher.password and rancher.autologin are in `ros config export`, but accessible as `ros config get`
	s.CheckOutput(c, "\n", Equals, "sudo ros config get rancher.password")
	s.CheckOutput(c,
		"EXTRA_CMDLINE: /init cc.hostname=nope rancher.password=three\n"+
			"    EXTRA_CMDLINE: /init cc.hostname=nope rancher.password=three\n"+
			"  password: three\n",
		Equals,
		"sudo ros config export | grep password",
	)

	// And then add a service.yml file example.
}
