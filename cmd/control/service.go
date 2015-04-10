package control

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

func serviceSubCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "enable",
			Usage:  "turn on an service",
			Action: enable,
		},
		{
			Name:   "disable",
			Usage:  "turn off an service",
			Action: disable,
		},
		{
			Name:   "list",
			Usage:  "list services and state",
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

	for _, service := range c.Args() {
		filtered := make([]string, 0, len(c.Args()))
		for _, existing := range cfg.EnabledServices {
			if existing != service {
				filtered = append(filtered, existing)
			}
		}

		if len(filtered) != len(c.Args()) {
			cfg.EnabledServices = filtered
			changed = true
		}
	}

	if changed {
		if err = cfg.Set("enabled_services", cfg.EnabledServices); err != nil {
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

	for _, service := range c.Args() {
		if !util.Contains(cfg.EnabledServices, service) {
			cfg.EnabledServices = append(cfg.EnabledServices, service)
			changed = true
		}
	}

	if changed {
		if err = cfg.Set("enabled_services", cfg.EnabledServices); err != nil {
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

	for _, service := range cfg.EnabledServices {
		enabled[service] = true
	}

	for service, _ := range cfg.Services {
		if _, ok := enabled[service]; ok {
			delete(enabled, service)
			fmt.Printf("enabled  %s\n", service)
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}

	for service, _ := range enabled {
		fmt.Printf("enabled  %s\n", service)
	}
}
