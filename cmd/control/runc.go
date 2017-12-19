package control

import (
	"fmt"

	"github.com/codegangsta/cli"

	"github.com/rancher/os/config"
	"github.com/rancher/os/init/runc"
	"github.com/rancher/os/util"
)

func runcCommand() cli.Command {
	var pivot cli.Flag
	if util.RootFsIsNotReal() {
		pivot = cli.BoolFlag{
			Name:  "pivot-root",
			Usage: "pivot-root (defaulted to false due to tmmpfs/ramfs)",
		}
	} else {
		pivot = cli.BoolTFlag{
			Name:  "pivot-root",
			Usage: "pivot-root (defaulted to true)",
		}
	}

	return cli.Command{
		Name:   "runc",
		Usage:  "create, prepare and run using runc",
		Action: runcAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "bundle, b",
				Usage: "path to the root of the bundle dir",
			},
			pivot,
			// TODO: add a --delete ?
		},
	}
}
func runcAction(c *cli.Context) error {
	fmt.Print("Runc start\n")
	serviceName := c.Args().Get(0)
	if serviceName == "" {
		fmt.Print("Please specify the service name (needs to be in the os-config)")
		return fmt.Errorf("Please specify the service name (needs to be in the os-config)")
	}
	bundleDir := c.String("bundle")
	pivotRoot := c.Bool("pivot-root")
	cfg := config.LoadConfig()
	return runc.Run(cfg, "", serviceName, bundleDir, !pivotRoot)
}
