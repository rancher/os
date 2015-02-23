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
	app.EnableBashCompletion = true

	app.Commands = []cli.Command{
		{
			Name:        "config",
			ShortName:   "c",
			Usage:       "configure settings",
			Subcommands: configSubcommands(),
		},
		{
			Name:        "addon",
			ShortName:   "a",
			Usage:       "addon settings",
			Subcommands: addonSubCommands(),
		},
		//{
		//	Name:      "reload",
		//	ShortName: "a",
		//	Usage:     "reload configuration of a service and restart the container",
		//	Action:    reload,
		//},
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
