package command

import (
	"errors"

	"github.com/codegangsta/cli"
	composeApp "github.com/docker/libcompose/cli/app"
	"github.com/rancher/os/cmd/control/service/app"
)

func verifyOneOrMoreServices(c *cli.Context) error {
	if len(c.Args()) == 0 {
		return errors.New("Must specify one or more services")
	}
	return nil
}

func CreateCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectCreate),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-recreate",
				Usage: "If containers already exist, don't recreate them. Incompatible with --force-recreate.",
			},
			cli.BoolFlag{
				Name:  "force-recreate",
				Usage: "Recreate containers even if their configuration and image haven't changed. Incompatible with --no-recreate.",
			},
			cli.BoolFlag{
				Name:  "no-build",
				Usage: "Don't build an image, even if it's missing.",
			},
		},
	}
}

func BuildCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "build",
		Usage:  "Build or rebuild services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectBuild),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "no-cache",
				Usage: "Do not use cache when building the image",
			},
			cli.BoolFlag{
				Name:  "force-rm",
				Usage: "Always remove intermediate containers",
			},
			cli.BoolFlag{
				Name:  "pull",
				Usage: "Always attempt to pull a newer version of the image",
			},
		},
	}
}

func PsCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "List containers",
		Action: composeApp.WithProject(factory, app.ProjectPs),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "q",
				Usage: "Only display IDs",
			},
		},
	}
}

func UpCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Create and start containers",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectUp),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "foreground",
				Usage: "Run in foreground and log",
			},
			cli.BoolFlag{
				Name:  "no-build",
				Usage: "Don't build an image, even if it's missing.",
			},
			cli.BoolFlag{
				Name:  "no-recreate",
				Usage: "If containers already exist, don't recreate them. Incompatible with --force-recreate.",
			},
			cli.BoolFlag{
				Name:  "force-recreate",
				Usage: "Recreate containers even if their configuration and image haven't changed. Incompatible with --no-recreate.",
			},
		},
	}
}

func StartCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "Start services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectStart),
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "foreground",
				Usage: "Run in foreground and log",
			},
		},
	}
}

func PullCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "pull",
		Usage:  "Pulls service images",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectPull),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "ignore-pull-failures",
				Usage: "Pull what it can and ignores images with pull failures.",
			},
		},
	}
}

func LogsCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "View output from containers",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectLog),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "lines",
				Usage: "number of lines to tail",
				Value: 100,
			},
			cli.BoolFlag{
				Name:  "follow",
				Usage: "Follow log output.",
			},
		},
	}
}

func RestartCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "restart",
		Usage:  "Restart services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectRestart),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

func StopCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "Stop services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectStop),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

func DownCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "Stop and remove containers, networks, images, and volumes",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectDown),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "volumes,v",
				Usage: "Remove data volumes",
			},
			cli.StringFlag{
				Name:  "rmi",
				Usage: "Remove images, type may be one of: 'all' to remove all images, or 'local' to remove only images that don't have an custom name set by the `image` field",
			},
			cli.BoolFlag{
				Name:  "remove-orphans",
				Usage: "Remove containers for services not defined in the Compose file",
			},
		},
	}
}

func RmCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "rm",
		Usage:  "Delete services",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectDelete),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force,f",
				Usage: "Allow deletion of all services",
			},
			cli.BoolFlag{
				Name:  "v",
				Usage: "Remove volumes associated with containers",
			},
		},
	}
}

func KillCommand(factory composeApp.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "kill",
		Usage:  "Kill containers",
		Before: verifyOneOrMoreServices,
		Action: composeApp.WithProject(factory, app.ProjectKill),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "signal,s",
				Usage: "SIGNAL to send to the container",
				Value: "SIGKILL",
			},
		},
	}
}
