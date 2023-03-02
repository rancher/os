package control

import (
	"fmt"
	"os"

	"github.com/burmilla/os/cmd/control/service"
	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/log"

	"github.com/codegangsta/cli"
)

func Main() {
	log.InitLogger()
	cli.VersionPrinter = func(c *cli.Context) {
		cfg := config.LoadConfig()
		runningName := cfg.Rancher.Upgrade.Image + ":" + config.Version
		fmt.Fprintf(c.App.Writer, "version %s from os image %s\n", c.App.Version, runningName)
	}
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("Control and configure BurmillaOS\nbuilt: %s", config.BuildDate)
	app.Version = config.Version
	app.Author = "Project Burmilla\n\tRancher Labs, Inc."
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
			Usage:       "manage which console container is used",
			HideHelp:    true,
			Subcommands: consoleSubcommands(),
		},
		{
			Name:            "console-init",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          consoleInitAction,
		},
		{
			Name:            "dev",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          devAction,
		},
		{
			Name:            "docker-init",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          dockerInitAction,
		},
		{
			Name:        "engine",
			Usage:       "manage which Docker engine is used",
			HideHelp:    true,
			Subcommands: engineSubcommands(),
		},
		{
			Name:            "entrypoint",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          entrypointAction,
		},
		{
			Name:            "env",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          envAction,
		},
		service.Commands(),
		{
			Name:        "os",
			Usage:       "operating system upgrade/downgrade",
			HideHelp:    true,
			Subcommands: osSubcommands(),
		},
		{
			Name:            "preload-images",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          preloadImagesAction,
		},
		{
			Name:            "recovery-init",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          recoveryInitAction,
		},
		{
			Name:            "switch-console",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          switchConsoleAction,
		},
		{
			Name:        "tls",
			Usage:       "setup tls configuration",
			HideHelp:    true,
			Subcommands: tlsConfCommands(),
		},
		{
			Name:            "udev-settle",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          udevSettleAction,
		},
		{
			Name:            "user-docker",
			Hidden:          true,
			HideHelp:        true,
			SkipFlagParsing: true,
			Action:          userDockerAction,
		},
		installCommand,
	}

	app.Run(os.Args)
}
