package control

import (
	"fmt"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/pkg/util"
)

func devAction(c *cli.Context) error {
	if len(c.Args()) > 0 {
		fmt.Println(util.ResolveDevice(c.Args()[0]))
	}
	return nil
}
