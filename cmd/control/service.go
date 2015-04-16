package control

import (
	"fmt"
	"log"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
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
		if _, ok := cfg.Services[service]; !ok {
			continue
		}

		cfg.Services[service] = false
		changed = true
	}

	if changed {
		if err = cfg.Set("services", cfg.Services); err != nil {
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
		if val, ok := cfg.Services[service]; !ok || !val {
			cfg.Services[service] = true
			changed = true
		}
	}

	if changed {
		if err = cfg.Set("services", cfg.Services); err != nil {
			log.Fatal(err)
		}
	}
}

func list(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for service, enabled := range cfg.Services {
		if enabled {
			fmt.Printf("enabled  %s\n", service)
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}
}
