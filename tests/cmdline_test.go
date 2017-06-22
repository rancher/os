package integration

import (
	. "gopkg.in/check.v1"

	"fmt"
	"strings"
)

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
	cmdline := s.CheckOutput(c, "", Not(Equals), "cat /proc/cmdline",)
	if strings.Contains(cmdline, extra) {
		c.Errorf("/proc/cmdline (%s) contains info that should be elided (%s)", cmdline, extra)
	}
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
	s.CheckCall(c,
		`echo 'test:
  image: alpine
  command: echo "tell me a secret ${EXTRA_CMDLINE}"
  labels:
    io.rancher.os.scope: system
  environment:
  - EXTRA_CMDLINE
' > test.yml`)
	s.CheckCall(c, "sudo mv test.yml /var/lib/rancher/conf/test.yml")
	s.CheckCall(c, "sudo ros service enable /var/lib/rancher/conf/test.yml")
	s.CheckCall(c, "sudo ros service up test")
	s.CheckOutput(c,
		"test_1 | tell me a secret /init cc.hostname=nope rancher.password=three\n",
		Equals,
		"sudo ros service logs test | grep secret",
	)

	// TODO: add a test showing we have the right password set
}
