package integration

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

func init() {
	Suite(&QemuSuite{
		runCommand: "../scripts/run",
		sshCommand: "../scripts/ssh",
		qemuCmd:    nil,
	})
}

var (
	BusyboxImage = map[string]string{
		"amd64": "busybox",
		"arm":   "armhf/busybox",
		"arm64": "aarch64/busybox",
	}[runtime.GOARCH]
	NginxImage = map[string]string{
		"amd64": "nginx",
		"arm":   "armhfbuild/nginx",
		"arm64": "armhfbuild/nginx",
	}[runtime.GOARCH]
	Version = os.Getenv("VERSION")
	Suffix  = os.Getenv("SUFFIX")
)

type QemuSuite struct {
	runCommand string
	sshCommand string
	qemuCmd    *exec.Cmd
}

func (s *QemuSuite) TearDownTest(c *C) {
	if s.qemuCmd != nil {
		s.Stop(c)
	}
}

// RunQemuWith requires user to specify all the `scripts/run` arguments
func (s *QemuSuite) RunQemuWith(c *C, additionalArgs ...string) error {

	err := s.runQemu(c, additionalArgs...)
	c.Assert(err, IsNil)
	return err
}

func (s *QemuSuite) RunQemu(c *C, additionalArgs ...string) error {
	runArgs := []string{
		"--qemu",
		"--no-rebuild",
		"--no-rm-usr",
		"--fresh",
	}
	runArgs = append(runArgs, additionalArgs...)

	err := s.RunQemuWith(c, runArgs...)
	c.Assert(err, IsNil)
	return err
}

func (s *QemuSuite) RunQemuInstalled(c *C, additionalArgs ...string) error {
	runArgs := []string{
		"--fresh",
	}
	runArgs = append(runArgs, additionalArgs...)

	err := s.RunQemu(c, runArgs...)
	c.Assert(err, IsNil)
	return err
}

func (s *QemuSuite) runQemu(c *C, args ...string) error {
	c.Assert(s.qemuCmd, IsNil) // can't run 2 qemu's at once (yet)
	s.qemuCmd = exec.Command(s.runCommand, args...)
	//s.qemuCmd.Stdout = os.Stdout
	s.qemuCmd.Stderr = os.Stderr
	if err := s.qemuCmd.Start(); err != nil {
		return err
	}
	fmt.Printf("--- %s: starting qemu %s, %v\n", c.TestName(), s.runCommand, args)

	return s.WaitForSSH()
}

func (s *QemuSuite) WaitForSSH() error {
	sshArgs := []string{
		"--qemu",
		"true",
	}

	var err error
	for i := 0; i < 100; i++ {
		cmd := exec.Command(s.sshCommand, sshArgs...)
		if err = cmd.Run(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		return fmt.Errorf("Failed to connect to SSH: %v", err)
	}

	sshArgs = []string{
		"--qemu",
		"docker",
		"version",
		">/dev/null",
		"2>&1",
	}

	for i := 0; i < 20; i++ {
		cmd := exec.Command(s.sshCommand, sshArgs...)
		if err = cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("Failed to check Docker version: %v", err)
}

func (s *QemuSuite) MakeCall(additionalArgs ...string) (string, error) {
	sshArgs := []string{
		"--qemu",
	}
	sshArgs = append(sshArgs, additionalArgs...)

	cmd := exec.Command(s.sshCommand, sshArgs...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	str := string(out)
	fmt.Println(str)
	return str, err
}

func (s *QemuSuite) CheckCall(c *C, additionalArgs ...string) {
	_, err := s.MakeCall(additionalArgs...)
	c.Assert(err, IsNil)
}

func (s *QemuSuite) CheckOutput(c *C, result string, check Checker, additionalArgs ...string) string {
	out, err := s.MakeCall(additionalArgs...)
	c.Assert(err, IsNil)
	c.Assert(out, check, result)
	return out
}

func (s *QemuSuite) Stop(c *C) {
	//s.MakeCall("sudo halt")
	//time.Sleep(2000 * time.Millisecond)
	//c.Assert(s.WaitForSSH(), IsNil)

	//fmt.Println("%s: stopping qemu", c.TestName())
	c.Assert(s.qemuCmd.Process.Kill(), IsNil)
	s.qemuCmd.Process.Wait()
	//time.Sleep(time.Millisecond * 1000)
	s.qemuCmd = nil
	fmt.Printf("--- %s: qemu stopped", c.TestName())
}

func (s *QemuSuite) Reboot(c *C) {
	fmt.Printf("--- %s: qemu reboot", c.TestName())
	s.MakeCall("sudo reboot")
	time.Sleep(3000 * time.Millisecond)
	c.Assert(s.WaitForSSH(), IsNil)
}

func (s *QemuSuite) LoadInstallerImage(c *C) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker save rancher/os:%s%s | ../scripts/ssh --qemu sudo system-docker load", Version, Suffix))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.Assert(cmd.Run(), IsNil)
}
