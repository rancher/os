package power

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
)

func Main() {
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Control and configure RancherOS"
	app.Version = config.VERSION
	app.Author = "Rancher Labs, Inc."
	app.Email = "sid@rancher.com"
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

func shutdown(c *cli.Context) {
	common("")
	reboot := c.String("r")
	poweroff := c.String("h")

	if reboot == "now" {
		Reboot()
	} else if poweroff == "now" {
		PowerOff()
	}
}
