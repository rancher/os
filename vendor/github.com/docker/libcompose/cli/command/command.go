package command

import (
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/cli/app"
)

// CreateCommand defines the libcompose create subcommand.
func CreateCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create all services but do not start",
		Action: app.WithProject(factory, app.ProjectCreate),
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

// BuildCommand defines the libcompose build subcommand.
func BuildCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "build",
		Usage:  "Build or rebuild services.",
		Action: app.WithProject(factory, app.ProjectBuild),
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

// PsCommand defines the libcompose ps subcommand.
func PsCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "List containers",
		Action: app.WithProject(factory, app.ProjectPs),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "q",
				Usage: "Only display IDs",
			},
		},
	}
}

// PortCommand defines the libcompose port subcommand.
func PortCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "port",
		Usage:  "Print the public port for a port binding",
		Action: app.WithProject(factory, app.ProjectPort),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "protocol",
				Usage: "tcp or udp ",
				Value: "tcp",
			},
			cli.IntFlag{
				Name:  "index",
				Usage: "index of the container if there are multiple instances of a service",
				Value: 1,
			},
		},
	}
}

// UpCommand defines the libcompose up subcommand.
func UpCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Bring all services up",
		Action: app.WithProject(factory, app.ProjectUp),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "d",
				Usage: "Do not block and log",
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

// StartCommand defines the libcompose start subcommand.
func StartCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "Start services",
		Action: app.WithProject(factory, app.ProjectStart),
		Flags: []cli.Flag{
			cli.BoolTFlag{
				Name:  "d",
				Usage: "Do not block and log",
			},
		},
	}
}

// RunCommand defines the libcompose run subcommand.
func RunCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "run",
		Usage:  "Run a one-off command",
		Action: app.WithProject(factory, app.ProjectRun),
		Flags:  []cli.Flag{},
	}
}

// PullCommand defines the libcompose pull subcommand.
func PullCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "pull",
		Usage:  "Pulls images for services",
		Action: app.WithProject(factory, app.ProjectPull),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "ignore-pull-failures",
				Usage: "Pull what it can and ignores images with pull failures.",
			},
		},
	}
}

// LogsCommand defines the libcompose logs subcommand.
func LogsCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "logs",
		Usage:  "Get service logs",
		Action: app.WithProject(factory, app.ProjectLog),
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

// RestartCommand defines the libcompose restart subcommand.
func RestartCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "restart",
		Usage:  "Restart services",
		Action: app.WithProject(factory, app.ProjectRestart),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

// StopCommand defines the libcompose stop subcommand.
func StopCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "Stop services",
		Action: app.WithProject(factory, app.ProjectStop),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

// DownCommand defines the libcompose stop subcommand.
func DownCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "Stop and remove containers, networks, images, and volumes",
		Action: app.WithProject(factory, app.ProjectDown),
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

// ScaleCommand defines the libcompose scale subcommand.
func ScaleCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "Scale services",
		Action: app.WithProject(factory, app.ProjectScale),
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "timeout,t",
				Usage: "Specify a shutdown timeout in seconds.",
				Value: 10,
			},
		},
	}
}

// RmCommand defines the libcompose rm subcommand.
func RmCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "rm",
		Usage:  "Delete services",
		Action: app.WithProject(factory, app.ProjectDelete),
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

// KillCommand defines the libcompose kill subcommand.
func KillCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "kill",
		Usage:  "Force stop service containers",
		Action: app.WithProject(factory, app.ProjectKill),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "signal,s",
				Usage: "SIGNAL to send to the container",
				Value: "SIGKILL",
			},
		},
	}
}

// PauseCommand defines the libcompose pause subcommand.
func PauseCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "pause",
		Usage: "Pause services.",
		// ArgsUsage: "[SERVICE...]",
		Action: app.WithProject(factory, app.ProjectPause),
	}
}

// UnpauseCommand defines the libcompose unpause subcommand.
func UnpauseCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "unpause",
		Usage: "Unpause services.",
		// ArgsUsage: "[SERVICE...]",
		Action: app.WithProject(factory, app.ProjectUnpause),
	}
}

// VersionCommand defines the libcompose version subcommand.
func VersionCommand(factory app.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "version",
		Usage:  "Show version informations",
		Action: app.Version,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "short",
				Usage: "Shows only Compose's version number.",
			},
		},
	}
}

// CommonFlags defines the flags that are in common for all subcommands.
func CommonFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name: "verbose,debug",
		},
		cli.StringSliceFlag{
			Name:   "file,f",
			Usage:  "Specify one or more alternate compose files (default: docker-compose.yml)",
			Value:  &cli.StringSlice{},
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:   "project-name,p",
			Usage:  "Specify an alternate project name (default: directory name)",
			EnvVar: "COMPOSE_PROJECT_NAME",
		},
	}
}
