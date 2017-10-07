package integration

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gbazil/telnet"

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
	netConsole telnet.Telnet
}

func (s *QemuSuite) TearDownTest(c *C) {
	if s.qemuCmd != nil {
		s.Stop(c)
	}
	time.Sleep(time.Second)
}

// RunQemuWith requires user to specify all the `scripts/run` arguments
func (s *QemuSuite) RunQemuWith(c *C, additionalArgs ...string) error {

	err := s.runQemu(c, additionalArgs...)
	c.Assert(err, IsNil)
	err = s.WaitForSSH()
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

// RunQemuWithNetConsole requires user to specify all the `scripts/run` arguments
func (s *QemuSuite) RunQemuWithNetConsole(c *C, additionalArgs ...string) error {
	runArgs := []string{
		"--netconsole",
	}
	runArgs = append(runArgs, additionalArgs...)

	err := s.runQemu(c, runArgs...)
	c.Assert(err, IsNil)

	time.Sleep(500 * time.Millisecond)
	// start telnet, and wait for prompt
	for i := 0; i < 20; i++ {
		s.netConsole, err = telnet.DialTimeout("127.0.0.1:4444", 5*time.Second)
		if err == nil {
			fmt.Printf("t%d SUCCEEDED\n", i)
			break
		}
		fmt.Printf("t%d", i)
		time.Sleep(500 * time.Millisecond)
	}
	c.Assert(err, IsNil)

	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)

		res := s.NetCall("uname")
		if strings.Contains(res, "Linux") {
			fmt.Printf("W%d SUCCEEDED(%s)\n", i, res)
			break
		}
	}

	s.NetCall("ip a")
	s.NetCall("cat /proc/cmdline")

	return err
}

func (s *QemuSuite) NetCall(cmd string) string {
	s.netConsole.Write(cmd + "\n")
	r, err := s.netConsole.Read("\n")
	fmt.Printf("cmd> %s", r)
	result := ""
	r = ""
	for err == nil {
		r, err = s.netConsole.Read("\n")
		fmt.Printf("\t%s", r)
		result = result + r
	}
	fmt.Printf("\n")
	// Note, if the result contains something like "+ cmd\n", you may have set -xe on
	return result
}
func (s *QemuSuite) NetCheckCall(c *C, additionalArgs ...string) {
	out := s.NetCall(strings.Join(additionalArgs, " "))
	c.Assert(out, Not(Equals), "")
}
func (s *QemuSuite) NetCheckOutput(c *C, result string, check Checker, additionalArgs ...string) string {
	out := s.NetCall(strings.Join(additionalArgs, " "))
	out = strings.Replace(out, "\r", "", -1)
	c.Assert(out, check, result)
	return out
}

func (s *QemuSuite) runQemu(c *C, args ...string) error {
	c.Assert(s.qemuCmd, IsNil) // can't run 2 qemu's at once (yet)
	time.Sleep(time.Second)
	s.qemuCmd = exec.Command(s.runCommand, args...)
	if os.Getenv("DEBUG") != "" {
		s.qemuCmd.Stdout = os.Stdout
		s.qemuCmd.Stderr = os.Stderr
	}
	if err := s.qemuCmd.Start(); err != nil {
		return err
	}
	fmt.Printf("--- %s: starting qemu %s, %v\n", c.TestName(), s.runCommand, args)

	return nil
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
		fmt.Printf("s%d", i)
		time.Sleep(time.Second)
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
		fmt.Printf("d%d", i)
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

func (s *QemuSuite) CheckOutputContains(c *C, result string, additionalArgs ...string) string {
	out, err := s.MakeCall(additionalArgs...)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(out, result), Equals, true)
	return out
}

func (s *QemuSuite) Stop(c *C) {
	fmt.Printf("%s: stopping qemu\n", c.TestName())
	//s.MakeCall("sudo poweroff")
	time.Sleep(1000 * time.Millisecond)
	//c.Assert(s.WaitForSSH(), IsNil)

	fmt.Printf("%s: stopping qemu 2\n", c.TestName())
	c.Assert(s.qemuCmd.Process.Kill(), IsNil)
	fmt.Printf("%s: stopping qemu 3\n", c.TestName())
	s.qemuCmd.Process.Wait()
	time.Sleep(time.Second)
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

func (s *QemuSuite) PullAndLoadImage(c *C, image string) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker pull %s", image))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.Assert(cmd.Run(), IsNil)

	cmd = exec.Command("sh", "-c", fmt.Sprintf("docker save %s | ../scripts/ssh --qemu sudo system-docker load", image))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.Assert(cmd.Run(), IsNil)
}
