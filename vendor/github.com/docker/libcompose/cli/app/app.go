package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

// ProjectAction is an adapter to allow the use of ordinary functions as libcompose actions.
// Any function that has the appropriate signature can be register as an action on a codegansta/cli command.
//
// cli.Command{
//		Name:   "ps",
//		Usage:  "List containers",
//		Action: app.WithProject(factory, app.ProjectPs),
//	}
type ProjectAction func(project *project.Project, c *cli.Context)

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.Warning("Note: This is an experimental alternate implementation of the Compose CLI (https://github.com/docker/compose)")
	return nil
}

// WithProject is an helper function to create a cli.Command action with a ProjectFactory.
func WithProject(factory ProjectFactory, action ProjectAction) func(context *cli.Context) {
	return func(context *cli.Context) {
		p, err := factory.Create(context)
		if err != nil {
			logrus.Fatalf("Failed to read project: %v", err)
		}
		action(p, context)
	}
}

// ProjectPs lists the containers.
func ProjectPs(p *project.Project, c *cli.Context) {
	allInfo := project.InfoSet{}
	qFlag := c.Bool("q")
	for name := range p.Configs {
		service, err := p.CreateService(name)
		if err != nil {
			logrus.Fatal(err)
		}

		info, err := service.Info(qFlag)
		if err != nil {
			logrus.Fatal(err)
		}

		allInfo = append(allInfo, info...)
	}

	os.Stdout.WriteString(allInfo.String(!qFlag))
}

// ProjectPort prints the public port for a port binding.
func ProjectPort(p *project.Project, c *cli.Context) {
	if len(c.Args()) != 2 {
		logrus.Fatalf("Please pass arguments in the form: SERVICE PORT")
	}

	index := c.Int("index")
	protocol := c.String("protocol")

	service, err := p.CreateService(c.Args()[0])
	if err != nil {
		logrus.Fatal(err)
	}

	containers, err := service.Containers()
	if err != nil {
		logrus.Fatal(err)
	}

	if index < 1 || index > len(containers) {
		logrus.Fatalf("Invalid index %d", index)
	}

	output, err := containers[index-1].Port(fmt.Sprintf("%s/%s", c.Args()[1], protocol))
	if err != nil {
		logrus.Fatal(err)
	}
	fmt.Println(output)
}

// ProjectDown brings all services down.
func ProjectDown(p *project.Project, c *cli.Context) {
	err := p.Down(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectBuild builds or rebuilds services.
func ProjectBuild(p *project.Project, c *cli.Context) {
	err := p.Build(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectCreate creates all services but do not start them.
func ProjectCreate(p *project.Project, c *cli.Context) {
	err := p.Create(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectUp brings all services up.
func ProjectUp(p *project.Project, c *cli.Context) {
	err := p.Up(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}

	if !c.Bool("d") {
		wait()
	}
}

// ProjectStart starts services.
func ProjectStart(p *project.Project, c *cli.Context) {
	err := p.Start(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectRestart restarts services.
func ProjectRestart(p *project.Project, c *cli.Context) {
	err := p.Restart(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectLog gets services logs.
func ProjectLog(p *project.Project, c *cli.Context) {
	err := p.Log(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
	wait()
}

// ProjectPull pulls images for services.
func ProjectPull(p *project.Project, c *cli.Context) {
	err := p.Pull(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectDelete delete services.
func ProjectDelete(p *project.Project, c *cli.Context) {
	if !c.Bool("force") && len(c.Args()) == 0 {
		logrus.Fatal("Will not remove all services without --force")
	}
	err := p.Delete(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectKill forces stop service containers.
func ProjectKill(p *project.Project, c *cli.Context) {
	err := p.Kill(c.Args()...)
	if err != nil {
		logrus.Fatal(err)
	}
}

// ProjectScale scales services.
func ProjectScale(p *project.Project, c *cli.Context) {
	// This code is a bit verbose but I wanted to parse everything up front
	order := make([]string, 0, 0)
	serviceScale := make(map[string]int)
	services := make(map[string]project.Service)

	for _, arg := range c.Args() {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			logrus.Fatalf("Invalid scale parameter: %s", arg)
		}

		name := kv[0]

		count, err := strconv.Atoi(kv[1])
		if err != nil {
			logrus.Fatalf("Invalid scale parameter: %v", err)
		}

		if _, ok := p.Configs[name]; !ok {
			logrus.Fatalf("%s is not defined in the template", name)
		}

		service, err := p.CreateService(name)
		if err != nil {
			logrus.Fatalf("Failed to lookup service: %s: %v", service, err)
		}

		order = append(order, name)
		serviceScale[name] = count
		services[name] = service
	}

	for _, name := range order {
		scale := serviceScale[name]
		logrus.Infof("Setting scale %s=%d...", name, scale)
		err := services[name].Scale(scale)
		if err != nil {
			logrus.Fatalf("Failed to set the scale %s=%d: %v", name, scale, err)
		}
	}
}

func wait() {
	<-make(chan interface{})
}
