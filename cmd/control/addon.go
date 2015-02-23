package control

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

func addonSubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "enable",
			Usage:  "turn on a module",
			Action: enable,
		},
		{
			Name:   "disable",
			Usage:  "turn off a module",
			Action: disable,
		},
		{
			Name:   "list",
			Usage:  "list modules and state",
			Action: list,
		},
	}
}

func disable(c *cli.Context) {
	changed := false
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for _, addon := range c.Args() {
		filtered := make([]string, 0, len(c.Args()))
		for _, existing := range cfg.EnabledAddons {
			if existing != addon {
				filtered = append(filtered, existing)
			}
		}

		if len(filtered) != len(c.Args()) {
			cfg.EnabledAddons = filtered
			changed = true
		}
	}

	if changed {
		if err = cfg.Save(); err != nil {
			log.Fatal(err)
		}
	}
}

func enable(c *cli.Context) {
	changed := false
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for _, addon := range c.Args() {
		if _, ok := cfg.Addons[addon]; ok && !util.Contains(cfg.EnabledAddons, addon) {
			cfg.EnabledAddons = append(cfg.EnabledAddons, addon)
			changed = true
		}
	}

	if changed {
		if err = cfg.Save(); err != nil {
			log.Fatal(err)
		}
	}
}

func list(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	enabled := map[string]bool{}

	for _, addon := range cfg.EnabledAddons {
		enabled[addon] = true
	}

	for addon, _ := range cfg.Addons {
		if _, ok := enabled[addon]; ok {
			fmt.Printf("%s enabled\n", addon)
		} else {
			fmt.Printf("%s disabled\n", addon)
		}
	}
}
