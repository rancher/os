package integration

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"testing"
	"time"

	. "github.com/cpuguy83/check"
)

func Test(t *testing.T) {
	TestingT(t)
}

func init() {
	Suite(&QemuSuite{
		runCommand: "../scripts/run",
		sshCommand: "../scripts/ssh",
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
	DockerUrl = "https://experimental.docker.com/builds/Linux/x86_64/docker-1.10.0-dev"
	Version   = os.Getenv("VERSION")
	Suffix    = os.Getenv("SUFFIX")
)

type QemuSuite struct {
	runCommand string
	sshCommand string
	qemuCmd    *exec.Cmd
}

// Sadly, getName will return "TearDownTest", so this doesn't know what test it was for.
func (s *QemuSuite) TearDownTest(c *C) {
}

func (s *QemuSuite) stopQemu(c *C) {
	runArgs := []string{
		"kill",
		getName(c),
	}
	fmt.Printf("Running %s %v\n", "docker", runArgs)
	s.qemuCmd = exec.Command("docker", runArgs...)
	s.qemuCmd.Stdout = os.Stdout
	s.qemuCmd.Stderr = os.Stderr
	if err := s.qemuCmd.Start(); err != nil {
		fmt.Printf("Error killing container %v\n", err)
	}

	time.Sleep(time.Millisecond * 1000)
}

func (s *QemuSuite) RunQemu(c *C, additionalArgs ...string) error {
	runArgs := []string{
		"--fresh",
	}
	runArgs = append(runArgs, additionalArgs...)

	return s.runQemu(c, runArgs...)
}

func (s *QemuSuite) RunQemuInstalled(c *C, additionalArgs ...string) error {
	runArgs := []string{
		"--installed",
	}
	runArgs = append(runArgs, additionalArgs...)

	return s.runQemu(c, runArgs...)
}

func (s *QemuSuite) runQemu(c *C, args ...string) error {
	runArgs := []string{
		"--qind",
		"--name",
		getName(c),
		"--no-rebuild",
		"--no-rm-usr",
	}
	runArgs = append(runArgs, args...)

	fmt.Printf("Running %s %v\n", s.runCommand, runArgs)
	s.qemuCmd = exec.Command(s.runCommand, runArgs...)
	s.qemuCmd.Stdout = os.Stdout
	s.qemuCmd.Stderr = os.Stderr
	if err := s.qemuCmd.Start(); err != nil {
		return err
	}

	fmt.Println("Pausing to let the VM start")
	time.Sleep(10 * time.Second)

	return s.WaitForSSH(c)
}

func getName(c *C) string {
	v := reflect.ValueOf(*c)
	method := v.FieldByName("method").Elem()
	info := method.FieldByName("Info")
	name := info.FieldByName("Name").String()
	return name
}

func (s *QemuSuite) WaitForSSH(c *C) error {
	sshArgs := []string{
		"--qind",
		"--notty",
		"--name",
		getName(c),
		"ls -la",
	}

	var err error
	for i := 0; i < 100; i++ {
		if err = RunStreaming(s.sshCommand, sshArgs...); err == nil {
			fmt.Printf("\t%d %v  OK\n", i, time.Now())
			break
		} else {
			fmt.Printf("\t%d %v err: %v\n", i, time.Now(), err)
		}
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("Failed to connect to SSH: %v", err)
	}

	sshArgs = []string{
		"--qind",
		"--notty",
		"--name",
		getName(c),
		"docker",
		"version",
		">/dev/null",
		"2>&1",
	}

	for i := 0; i < 50; i++ {
		if err = RunStreaming(s.sshCommand, sshArgs...); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
		fmt.Printf("\t%d %v err: %v\n", i, time.Now(), err)
	}

	return fmt.Errorf("Failed to check Docker version: %v", err)
}

func RunStreaming(command string, args ...string) error {
//	fmt.Printf("Running %s\n", command)
//	for i, v := range args {
//		fmt.Printf("%d\t (%s)\n", i, v)
//	}

        cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
                return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
                return err
	}
        defer func() {
                _ = stdout.Close()
                _ = stderr.Close()
        }()

        err = cmd.Start()
        if err != nil {
		fmt.Println(err)
                return err
        }

        errscanner := bufio.NewScanner(stderr)
        go func() {
                for errscanner.Scan() {
                        fmt.Println(errscanner.Text())
                }
        }()
        outscanner := bufio.NewScanner(stdout)
        for outscanner.Scan() {
                str := outscanner.Text()
                fmt.Println(str)
        }
        if err := outscanner.Err(); err != nil {
		fmt.Println(err)
                return err
        }
        if err := cmd.Wait(); err != nil {
		fmt.Println(err)
                return err
        }
	return nil
}


func (s *QemuSuite) MakeCall(c *C, additionalArgs ...string) error {
	sshArgs := []string{
		"--qind",
		"--notty",
		"--name",
		getName(c),
	}
	sshArgs = append(sshArgs, additionalArgs...)

	return RunStreaming(s.sshCommand, sshArgs...)
}

func (s *QemuSuite) CheckCall(c *C, additionalArgs ...string) {
	c.Assert(s.MakeCall(c, additionalArgs...), IsNil)
}

func (s *QemuSuite) Reboot(c *C) {
	s.MakeCall(c, "sudo reboot")
	time.Sleep(3000 * time.Millisecond)
	c.Assert(s.WaitForSSH(c), IsNil)
}

func (s *QemuSuite) LoadInstallerImage(c *C) {
	fmt.Printf("Running %s %v\n", "sh", "-c", fmt.Sprintf("docker save rancher/os:%s%s | ../scripts/ssh --qind sudo system-docker load", Version, Suffix))
	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker save rancher/os:%s%s | ../scripts/ssh --qind sudo system-docker load", Version, Suffix))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	c.Assert(cmd.Run(), IsNil)
}
