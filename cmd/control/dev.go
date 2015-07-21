package control

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/rancherio/os/util"
)

func devAction(c *cli.Context) {
	fmt.Println(util.ResolveDevice(c.Args()[0]))
}
