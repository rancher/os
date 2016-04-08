package control

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/cli/command"
	dockerApp "github.com/docker/libcompose/cli/docker/app"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

type projectFactory struct {
}

func (p *projectFactory) Create(c *cli.Context) (*project.Project, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return compose.GetProject(cfg, true)
}

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func serviceCommand() cli.Command {
	factory := &projectFactory{}

	app := cli.Command{}
	app.Name = "service"
	app.ShortName = "s"
	app.Usage = "Command line interface for services and compose."
	app.Before = beforeApp
	app.Flags = append(command.CommonFlags(), dockerApp.DockerClientFlags()...)
	app.Subcommands = append(serviceSubCommands(),
		command.BuildCommand(factory),
		command.CreateCommand(factory),
		command.UpCommand(factory),
		command.StartCommand(factory),
		command.LogsCommand(factory),
		command.RestartCommand(factory),
		command.StopCommand(factory),
		command.ScaleCommand(factory),
		command.RmCommand(factory),
		command.PullCommand(factory),
		command.KillCommand(factory),
		command.PortCommand(factory),
		command.PsCommand(factory),
	)

	return app
}

func serviceSubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "enable",
			Usage:  "turn on an service",
			Action: enable,
		},
		{
			Name:   "disable",
			Usage:  "turn off an service",
			Action: disable,
		},
		{
			Name:   "list",
			Usage:  "list services and state",
			Action: list,
		},
		{
			Name:   "delete",
			Usage:  "delete a service",
			Action: del,
		},
	}
}

func disable(c *cli.Context) {
	changed := false
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	for _, service := range c.Args() {
		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			logrus.Fatalf("ERROR: Service %s is not a valid service, use \"sudo ros service list\" to list the services", service)
		}

		cfg.Rancher.ServicesInclude[service] = false
		changed = true
	}

	if changed {
		if err = cfg.Save(); err != nil {
			logrus.Fatal(err)
		}
	}
}

func del(c *cli.Context) {
	changed := false
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	for _, service := range c.Args() {
		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			logrus.Fatalf("ERROR: Service %s is not a valid service, use \"sudo ros service list\" to list the services", service)
		}
		delete(cfg.Rancher.ServicesInclude, service)
		changed = true
	}

	if changed {
		if err = cfg.Save(); err != nil {
			logrus.Fatal(err)
		}
	}
}

func enable(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	services, err := util.GetServices(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		logrus.Fatalf("Failed to get services: %v", err)
	}

	var enabledServices []string

	for _, service := range c.Args() {
		if val, ok := cfg.Rancher.ServicesInclude[service]; !ok || !val {
			if strings.HasPrefix(service, "/") && !strings.HasPrefix(service, "/var/lib/rancher/conf") {
				logrus.Fatalf("ERROR: Service should be in path /var/lib/rancher/conf")
			}

			if !util.Contains(services, service) {
				logrus.Fatalf("ERROR: Service %s is not a valid service, use \"sudo ros service list\" to list the services", service)
			}

			cfg.Rancher.ServicesInclude[service] = true
			enabledServices = append(enabledServices, service)
		}
	}

	if len(enabledServices) > 0 {
		if err := compose.StageServices(cfg, enabledServices...); err != nil {
			logrus.Fatal(err)
		}

		if err := cfg.Save(); err != nil {
			logrus.Fatal(err)
		}
	}
}

func list(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	clone := make(map[string]bool)
	for service, enabled := range cfg.Rancher.ServicesInclude {
		clone[service] = enabled
	}

	services, err := util.GetServices(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		logrus.Fatalf("Failed to get services: %v", err)
	}

	for _, service := range services {
		if enabled, ok := clone[service]; ok {
			delete(clone, service)
			if enabled {
				fmt.Printf("enabled  %s\n", service)
			} else {
				fmt.Printf("disabled %s\n", service)
			}
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}

	for service, enabled := range clone {
		if enabled {
			fmt.Printf("enabled  %s\n", service)
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}
}
