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
	app.Email = "darren@rancher.com"

	app.Commands = []cli.Command{
		{
			Name:        "config",
			ShortName:   "c",
			Usage:       "configure settings",
			Subcommands: configSubcommands(),
		},
		{
			Name:      "module",
			ShortName: "m",
			Usage:     "module settings",
			Subcommands: []cli.Command{
				{
					Name:  "activate",
					Usage: "turn on a module and possibly reboot",
				},
				{
					Name:  "deactivate",
					Usage: "turn off a module and possibly reboot",
				},
				{
					Name:  "list",
					Usage: "list modules and state",
				},
			},
		},
		{
			Name:  "os",
			Usage: "operating system upgrade/downgrade",
			Subcommands: osSubcommands(),
		},
		{
			Name: "tlsconf",
			Usage: "setup tls configuration",
			Subcommands: tlsConfCommands(),
		},
	}

	app.Run(os.Args)
}
