package service

import (
	"fmt"
	"strings"

	"github.com/codegangsta/cli"
	dockerApp "github.com/docker/libcompose/cli/docker/app"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/cmd/control/service/command"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

type projectFactory struct {
}

func (p *projectFactory) Create(c *cli.Context) (project.APIProject, error) {
	cfg := config.LoadConfig()
	return compose.GetProject(cfg, true, false)
}

func beforeApp(c *cli.Context) error {
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}

func Commands() cli.Command {
	factory := &projectFactory{}

	app := cli.Command{}
	app.Name = "service"
	app.ShortName = "s"
	app.Usage = "Command line interface for services and compose."
	app.Before = beforeApp
	app.Flags = append(dockerApp.DockerClientFlags(), cli.BoolFlag{
		Name: "verbose,debug",
	})
	app.Subcommands = append(serviceSubCommands(),
		command.BuildCommand(factory),
		command.CreateCommand(factory),
		command.UpCommand(factory),
		command.StartCommand(factory),
		command.LogsCommand(factory),
		command.RestartCommand(factory),
		command.StopCommand(factory),
		command.RmCommand(factory),
		command.PullCommand(factory),
		command.KillCommand(factory),
		command.PsCommand(factory),
	)

	return app
}

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

func updateIncludedServices(cfg *config.CloudConfig) error {
	return config.Set("rancher.services_include", cfg.Rancher.ServicesInclude)
}

func disable(c *cli.Context) error {
	changed := false
	cfg := config.LoadConfig()

	for _, service := range c.Args() {
		validateService(service, cfg)

		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			continue
		}

		cfg.Rancher.ServicesInclude[service] = false
		changed = true
	}

	if changed {
		if err := updateIncludedServices(cfg); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func del(c *cli.Context) error {
	changed := false
	cfg := config.LoadConfig()

	for _, service := range c.Args() {
		validateService(service, cfg)

		if _, ok := cfg.Rancher.ServicesInclude[service]; !ok {
			continue
		}

		delete(cfg.Rancher.ServicesInclude, service)
		changed = true
	}

	if changed {
		if err := updateIncludedServices(cfg); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func enable(c *cli.Context) error {
	cfg := config.LoadConfig()

	var enabledServices []string

	for _, service := range c.Args() {
		validateService(service, cfg)

		if val, ok := cfg.Rancher.ServicesInclude[service]; !ok || !val {
			if isLocal(service) && !strings.HasPrefix(service, "/var/lib/rancher/conf") {
				log.Fatalf("ERROR: Service should be in path /var/lib/rancher/conf")
			}

			cfg.Rancher.ServicesInclude[service] = true
			enabledServices = append(enabledServices, service)
		}
	}

	if len(enabledServices) > 0 {
		if err := compose.StageServices(cfg, enabledServices...); err != nil {
			log.Fatal(err)
		}

		if err := updateIncludedServices(cfg); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func list(c *cli.Context) error {
	cfg := config.LoadConfig()

	clone := make(map[string]bool)
	for service, enabled := range cfg.Rancher.ServicesInclude {
		clone[service] = enabled
	}

	services := availableService(cfg)

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

	return nil
}

func isLocal(service string) bool {
	return strings.HasPrefix(service, "/")
}

func IsLocalOrURL(service string) bool {
	return isLocal(service) || strings.HasPrefix(service, "http:/") || strings.HasPrefix(service, "https:/")
}

// Check to see if the service definition exists
func ValidService(service string, cfg *config.CloudConfig) bool {
	services := availableService(cfg)
	if !IsLocalOrURL(service) && !util.Contains(services, service) {
		return false
	}
	return true
}

func validateService(service string, cfg *config.CloudConfig) {
	if !ValidService(service, cfg) {
		log.Fatalf("%s is not a valid service", service)
	}
}

func availableService(cfg *config.CloudConfig) []string {
	services, err := network.GetServices(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		log.Fatalf("Failed to get services: %v", err)
	}
	return services
}
