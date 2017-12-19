package integration

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/net/context"

	. "gopkg.in/check.v1"
)

func (s *CliSuite) TestBuild(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build")
	err := cmd.Run()

	oneImageName := fmt.Sprintf("%s_one", p)
	twoImageName := fmt.Sprintf("%s_two", p)

	c.Assert(err, IsNil)

	client := GetClient(c)
	one, _, err := client.ImageInspectWithRaw(context.Background(), oneImageName, false)
	c.Assert(err, IsNil)
	c.Assert([]string(one.Config.Cmd), DeepEquals, []string{"echo", "one"})

	two, _, err := client.ImageInspectWithRaw(context.Background(), twoImageName, false)
	c.Assert(err, IsNil)
	c.Assert([]string(two.Config.Cmd), DeepEquals, []string{"echo", "two"})
}

func (s *CliSuite) TestBuildWithNoCache1(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build")

	output, err := cmd.Output()
	c.Assert(err, IsNil)

	cmd = exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build")
	output, err = cmd.Output()
	c.Assert(err, IsNil)
	out := string(output[:])
	c.Assert(strings.Contains(out,
		"Using cache"),
		Equals, true, Commentf("%s", out))
}

func (s *CliSuite) TestBuildWithNoCache2(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build")

	output, err := cmd.Output()
	c.Assert(err, IsNil)

	cmd = exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build", "--no-cache")
	output, err = cmd.Output()
	c.Assert(err, IsNil)
	out := string(output[:])
	c.Assert(strings.Contains(out,
		"Using cache"),
		Equals, false, Commentf("%s", out))
}

func (s *CliSuite) TestBuildWithNoCache3(c *C) {
	p := s.RandomProject()
	cmd := exec.Command(s.command, "-f", "./assets/build/docker-compose.yml", "-p", p, "build", "--no-cache")
	err := cmd.Run()

	oneImageName := fmt.Sprintf("%s_one", p)
	twoImageName := fmt.Sprintf("%s_two", p)

	c.Assert(err, IsNil)

	client := GetClient(c)
	one, _, err := client.ImageInspectWithRaw(context.Background(), oneImageName, false)
	c.Assert(err, IsNil)
	c.Assert([]string(one.Config.Cmd), DeepEquals, []string{"echo", "one"})

	two, _, err := client.ImageInspectWithRaw(context.Background(), twoImageName, false)
	c.Assert(err, IsNil)
	c.Assert([]string(two.Config.Cmd), DeepEquals, []string{"echo", "two"})
}
