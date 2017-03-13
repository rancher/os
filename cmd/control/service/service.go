package service

import (
	"context"
	"fmt"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/codegangsta/cli"
	dockerApp "github.com/docker/libcompose/cli/docker/app"
	composeConfig "github.com/docker/libcompose/config"

	"github.com/docker/libcompose/project/options"

	"github.com/docker/libcompose/project"
	"github.com/rancher/os/cmd/control/service/command"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

type ProjectFactory struct {
}

func (p *ProjectFactory) Create(c *cli.Context) (project.APIProject, error) {
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
	factory := &ProjectFactory{}

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
			Action: Enable,
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
			Action: Del,
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

func Del(c *cli.Context) error {
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

func ComposeToCloudConfig(bytes []byte) ([]byte, error) {
	//TODO: copied from cloudinitsave, move to config.
	compose := make(map[interface{}]interface{})
	err := yaml.Unmarshal(bytes, &compose)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"services": compose,
		},
	})
}

func LoadService(repoName, serviceLongName string) (*config.CloudConfig, error) {
	// TODO: this should move to something like config/service.go?
	// WARNING: this can contain more than one service - Josh and I aren't sure this is worth it
	servicePath := fmt.Sprintf("%s/%s.yml", repoName, serviceLongName)
	//log.Infof("loading %s", serviceLongName)
	content, err := network.CacheLookup(servicePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load %s: %v", servicePath, err)
	}
	if content, err = ComposeToCloudConfig(content); err != nil {
		return nil, fmt.Errorf("Failed to convert compose to cloud-config syntax: %v", err)
	}

	p, err := config.ReadConfig(content, true)
	if err != nil {
		return nil, fmt.Errorf("Failed to load %s : %v", servicePath, err)
	}
	return p, nil
}

func IsConsole(serviceCfg *config.CloudConfig) bool {
	// TODO: this should move to something like config/service.go?
	//the service is called console, and has an io.rancher.os.console label.
	for serviceName, service := range serviceCfg.Rancher.Services {
		if serviceName == "console" {
			for k := range service.Labels {
				if k == "io.rancher.os.console" {
					return true
				}
			}
		}
	}
	return false
}

func IsEngine(serviceCfg *config.CloudConfig) bool {
	// TODO: this should move to something like config/service.go?
	//the service is called docker, and the command is "ros user-docker"
	for serviceName, service := range serviceCfg.Rancher.Services {
		log.Infof("serviceName == %s", serviceName)
		if serviceName == "docker" {
			cmd := strings.Join(service.Command, " ")
			log.Infof("service command == %s", cmd)
			if cmd == "ros user-docker" {
				log.Infof("yes, its a docker engine")
				return true
			}
		}
	}
	return false
}

func Enable(c *cli.Context) error {
	cfg := config.LoadConfig()

	var enabledServices []string
	var consoleService, engineService string
	var errorServices []string
	serviceMap := make(map[string]*config.CloudConfig)

	for _, service := range c.Args() {
		//validateService(service, cfg)
		//log.Infof("start4")
		// TODO: need to search for the service in all the repos.
		// TODO: also need to deal with local file paths and URLs
		serviceConfig, err := LoadService("core", service)
		if err != nil {
			log.Errorf("Failed to load %s: %s", service, err)
			errorServices = append(errorServices, service)
			continue
		}
		serviceMap[service] = serviceConfig
	}
	if len(serviceMap) == 0 {
		log.Fatalf("No valid Services found")
	}
	if len(errorServices) > 0 {
		if c.Bool("force") || !util.Yes("Some services failed to load, Continue?") {
			log.Fatalf("Services failed to load: %v", errorServices)
		}
	}

	for service, serviceConfig := range serviceMap {
		if val, ok := cfg.Rancher.ServicesInclude[service]; !ok || !val {
			if isLocal(service) && !strings.HasPrefix(service, "/var/lib/rancher/conf") {
				log.Fatalf("ERROR: Service should be in path /var/lib/rancher/conf")
			}

			if IsConsole(serviceConfig) {
				log.Debugf("Enabling the %s console", service)
				if err := config.Set("rancher.console", service); err != nil {
					log.Errorf("Failed to update 'rancher.console': %v", err)
				}
				consoleService = service

			} else if IsEngine(serviceConfig) {
				log.Debugf("Enabling the %s user engine", service)
				if err := config.Set("rancher.docker.engine", service); err != nil {
					log.Errorf("Failed to update 'rancher.docker.engine': %v", err)
				}
				engineService = service
			} else {
				cfg.Rancher.ServicesInclude[service] = true
			}
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

	//TODO: fix the case where the user is applying both a new console and a new docker engine
	if consoleService != "" && c.Bool("apply") {
		//ros console switch.
		if !c.Bool("force") {
			fmt.Println(`Switching consoles will
1. destroy the current console container
2. log you out
3. restart Docker`)
			if !util.Yes("Continue") {
				return nil
			}
			switchService, err := compose.CreateService(nil, "switch-console", &composeConfig.ServiceConfigV1{
				LogDriver:  "json-file",
				Privileged: true,
				Net:        "host",
				Pid:        "host",
				Image:      config.OsBase,
				Labels: map[string]string{
					config.ScopeLabel: config.System,
				},
				Command:     []string{"/usr/bin/ros", "switch-console", consoleService},
				VolumesFrom: []string{"all-volumes"},
			})
			if err != nil {
				log.Fatal(err)
			}

			if err = switchService.Delete(context.Background(), options.Delete{}); err != nil {
				log.Fatal(err)
			}
			if err = switchService.Up(context.Background(), options.Up{}); err != nil {
				log.Fatal(err)
			}
			if err = switchService.Log(context.Background(), true); err != nil {
				log.Fatal(err)
			}
		}
	}
	if engineService != "" && c.Bool("apply") {
		log.Info("Starting the %s engine", engineService)
		project, err := compose.GetProject(cfg, true, false)
		if err != nil {
			log.Fatal(err)
		}

		if err = project.Stop(context.Background(), 10, "docker"); err != nil {
			log.Fatal(err)
		}

		if err = compose.LoadSpecialService(project, cfg, "docker", engineService); err != nil {
			log.Fatal(err)
		}

		if err = project.Up(context.Background(), options.Up{}, "docker"); err != nil {
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

func validateService(service string, cfg *config.CloudConfig) {
	services := availableService(cfg)
	if !IsLocalOrURL(service) && !util.Contains(services, service) {
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
