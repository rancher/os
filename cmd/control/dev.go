package control

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/util"
)

func devAction(c *cli.Context) {
	if len(c.Args()) > 0 {
		fmt.Println(util.ResolveDevice(c.Args()[0]))
	}
}
