package control

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/compose"
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
		{
			Name:   "delete",
			Usage:  "delete a service",
			Action: del,
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
		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			continue
		}

		cfg.Rancher.ServicesInclude[service] = false
		changed = true
	}

	if changed {
		if err = cfg.Set("rancher.services_include", cfg.Rancher.ServicesInclude); err != nil {
			log.Fatal(err)
		}
	}
}

func del(c *cli.Context) {
	changed := false
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for _, service := range c.Args() {
		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			continue
		}
		delete(cfg.Rancher.ServicesInclude, service)
		changed = true
	}

	if changed {
		if err = cfg.Set("rancher.services_include", cfg.Rancher.ServicesInclude); err != nil {
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
		if val, ok := cfg.Rancher.ServicesInclude[service]; !ok || !val {
			if strings.HasPrefix(service, "/") && !strings.HasPrefix(service, "/var/lib/rancher/conf") {
				log.Fatalf("ERROR: Service should be in path /var/lib/rancher/conf")
			}
			if _, err := compose.LoadServiceResource(service, true, cfg); err != nil {
				log.Fatalf("could not load service %s", service)
			}
			cfg.Rancher.ServicesInclude[service] = true
			changed = true
		}
	}

	if changed {
		if err := cfg.Set("rancher.services_include", cfg.Rancher.ServicesInclude); err != nil {
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
	for service, enabled := range cfg.Rancher.ServicesInclude {
		clone[service] = enabled
	}

	services, err := util.GetServices(cfg.Rancher.Repositories.ToArray())
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
