package control

import (
	"fmt"
	"log"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
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
		if _, ok := cfg.ServicesInclude[service]; !ok {
			continue
		}

		cfg.ServicesInclude[service] = false
		changed = true
	}

	if changed {
		if err = cfg.Set("services_include", cfg.ServicesInclude); err != nil {
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
		if val, ok := cfg.ServicesInclude[service]; !ok || !val {
			if strings.HasPrefix(service, "/") && !strings.HasPrefix(service, "/var/lib/rancher/conf") {
				log.Printf("ERROR: Service should be in path /var/lib/rancher/conf")
				continue
			}
			if err := docker.StageService(service, cfg); err != nil {
				log.Printf("Could not stage service %s", service)
				continue
			}
			cfg.ServicesInclude[service] = true
			changed = true
		}
	}

	if changed {
		if err = cfg.Set("services_include", cfg.ServicesInclude); err != nil {
			log.Fatal(err)
		}
	}
}

func list(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	clone := make(map[string]bool)
	for service, enabled := range cfg.ServicesInclude {
		clone[service] = enabled
	}

	services, err := util.GetServices(cfg.Repositories.ToArray())
	if err != nil {
		log.Fatalf("Failed to get services: %v", err)
	}

	for _, service := range services {
		if enabled, ok := clone[service]; ok {
			delete(clone, service)
			if enabled {
				fmt.Printf("enabled  %s\n", service)
			} else {
				fmt.Printf("disabled %s\n", service)
			}
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}

	for service, enabled := range clone {
		if enabled {
			fmt.Printf("enabled  %s\n", service)
		} else {
			fmt.Printf("disabled %s\n", service)
		}
	}
}
