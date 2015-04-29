package control

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
)

func Main() {
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Control and configure RancherOS"
	app.Version = config.VERSION
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:        "config",
			ShortName:   "c",
			Usage:       "configure settings",
			HideHelp:    true,
			Subcommands: configSubcommands(),
		},
		{
			Name:               "env",
			ShortName:          "e",
			Usage:              "env command",
			HideHelp:           true,
			SkipFlagParsing:    true,
			Action:             envAction,
		},
		{
			Name:        "service",
			ShortName:   "s",
			Usage:       "service settings",
			HideHelp:    true,
			Subcommands: serviceSubCommands(),
		},
		//{
		//	Name:      "reload",
		//	ShortName: "a",
		//	Usage:     "reload configuration of a service and restart the container",
		//	Action:    reload,
		//},
		{
			Name:        "os",
			Usage:       "operating system upgrade/downgrade",
			HideHelp:    true,
			Subcommands: osSubcommands(),
		},
		{
			Name:        "tls",
			Usage:       "setup tls configuration",
			HideHelp:    true,
			Subcommands: tlsConfCommands(),
		},
	}

	app.Run(os.Args)
}
