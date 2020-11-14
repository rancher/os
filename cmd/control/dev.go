package control

import (
	"fmt"

	"github.com/burmilla/os/pkg/util"

	"github.com/codegangsta/cli"
)

func devAction(c *cli.Context) error {
	if len(c.Args()) > 0 {
		fmt.Println(util.ResolveDevice(c.Args()[0]))
	}
	return nil
}
