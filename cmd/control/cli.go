package control

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/cmd/control/service"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func Main() {
	log.InitLogger()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("Control and configure RancherOS\nbuilt: %s", config.BuildDate)
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Before = func(c *cli.Context) error {
		if os.Geteuid() != 0 &&
			filepath.Base(os.Args[0]) != "host_ros" {
			log.Fatalf("%s: Need to be root", os.Args[0])
		}
		return nil
	}

	app.Commands = []cli.Command{
//		{
//			Name:            "bootstrap",
//			Hidden:          true,
//			HideHelp:        true,
//			SkipFlagParsing: true,
//			Action:          bootstrapAction,
//		},
		runcCommand(),
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
		selinuxCommand(),
	}

	app.Run(os.Args)
}
