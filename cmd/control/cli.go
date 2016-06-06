package control

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
)

func Main() {
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Control and configure RancherOS"
	app.Version = config.VERSION
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Before = func(c *cli.Context) error {
		if os.Geteuid() != 0 {
			log.Fatalf("%s: Need to be root", os.Args[0])
		}
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:        "config",
			ShortName:   "c",
			Usage:       "configure settings",
			HideHelp:    true,
			Subcommands: configSubcommands(),
		},
		{
			Name:        "console",
			Usage:       "console container commands",
			HideHelp:    true,
			Subcommands: consoleSubcommands(),
		},
		{
			Name:            "dev",
			ShortName:       "d",
			Usage:           "dev spec",
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          devAction,
		},
		{
			Name:            "env",
			ShortName:       "e",
			Usage:           "env command",
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          envAction,
		},
		serviceCommand(),
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
		installCommand,
		selinuxCommand(),
	}

	app.Run(os.Args)
}
