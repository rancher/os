package control

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/cli/command"
	dockerApp "github.com/docker/libcompose/cli/docker/app"
	"github.com/docker/libcompose/project"
	"github.com/rancherio/os/compose"
	"github.com/rancherio/os/config"
)

type projectFactory struct {
}

func (p *projectFactory) Create(c *cli.Context) (*project.Project, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	return compose.GetProject(cfg)
}

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	}
	return nil
}

func composeCommand() cli.Command {
	factory := &projectFactory{}

	app := cli.Command{}
	app.Name = "compose"
	app.Usage = "Command line interface for libcompose."
	app.Before = beforeApp
	app.Flags = append(command.CommonFlags(), dockerApp.DockerClientFlags()...)
	app.Subcommands = []cli.Command{
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
	}

	return app
}
