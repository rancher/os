package power

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func Main() {
	log.InitLogger()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("%s RancherOS\nbuilt: %s", app.Name, config.BuildDate)
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Action = shutdown
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "r, R",
			Usage: "reboot after shutdown",
		},
		cli.StringFlag{
			Name:  "h",
			Usage: "halt the system",
		},
	}
	app.HideHelp = true

	app.Run(os.Args)
}

func shutdown(c *cli.Context) error {
	common("")
	reboot := c.String("r")
	poweroff := c.String("h")

	if reboot == "now" {
		Reboot()
	} else if poweroff == "now" {
		Off()
	}

	return nil
}
