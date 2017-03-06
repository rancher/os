package control

import (
	"fmt"
	"os"
	"sort"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/codegangsta/cli"
	composeApp "github.com/docker/libcompose/cli/app"
	libcomposeConfig "github.com/docker/libcompose/config"
	serviceApp "github.com/rancher/os/cmd/control/service/app"

	"github.com/rancher/os/cmd/control/service"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util/network"
)

func Main() {
	log.InitLogger()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Control and configure RancherOS"
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Before = func(c *cli.Context) error {
		if os.Geteuid() != 0 {
			log.Fatalf("%s: Need to be root", os.Args[0])
		}
		return nil
	}

	factory := &service.ProjectFactory{}

	app.Commands = []cli.Command{
		{
			Name:      "fetch",
			ShortName: "f",
			Usage:     "fetch configs from repos",
			HideHelp:  true,
			Action:    fetchServices,
		},
		// service
		{
			Name:      "list",
			ShortName: "",
			Usage:     "list services and states",
			HideHelp:  true,
			Action:    listServices,
		}, {
			Name: "install",
			// TODO: add an --apply or --up ...
			// TODO: also support the repo-name prefix
			ShortName: "",
			Usage:     "install/upgrade service / RancherOS",
			HideHelp:  true,
			Action:    service.Enable,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "apply",
					Usage: "Switch console/engine, or start service.",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "Don't ask questions.",
				},
			},
		}, {
			Name:      "remove",
			ShortName: "",
			Usage:     "remove service",
			HideHelp:  true,
			Action:    service.Del,
		}, {
			Name:  "logs",
			Usage: "View output from containers",
			//Before: verifyOneOrMoreServices,
			Action: composeApp.WithProject(factory, serviceApp.ProjectLog),
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "lines",
					Usage: "number of lines to tail",
					Value: 100,
				},
				cli.BoolFlag{
					Name:  "follow",
					Usage: "Follow log output.",
				},
			},
		},
		// settings / partial configs
		{
			Name: "get",
			// TODO: also add the merge command functionality
			ShortName: "",
			Usage:     "get config value(s)",
			HideHelp:  true,
			Action:    configGet,
		}, {
			Name:      "set",
			ShortName: "",
			Usage:     "set config value(s)",
			HideHelp:  true,
			Action:    configSet,
		},
		// complete config
		{
			Name:  "export",
			Usage: "export configuration",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "File to which to save",
				},
				cli.BoolFlag{
					Name:  "private, p",
					Usage: "Include the generated private keys",
				},
				cli.BoolFlag{
					Name:  "full, f",
					Usage: "Export full configuration, including internal and default settings",
				},
			},
			Action: export,
		}, {
			Name:   "validate",
			Usage:  "validate configuration from stdin",
			Action: validate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read",
				},
			},
		}, {
			Name:      "apply",
			ShortName: "",
			Usage:     "apply service&config changes",
			HideHelp:  true,
			Action:    dummy,
		},
		// old..
		{
			Name:        "old",
			ShortName:   "o",
			Usage:       "old Command line (deprecated, will be removed in future)",
			HideHelp:    true,
			Subcommands: originalCli,
		},
	}
	app.Commands = append(app.Commands, hiddenInternalCommands...)

	app.Run(os.Args)
}

func dummy(c *cli.Context) error {
	fmt.Printf("Not implemented yet - use the `ros old` commands for now\n")
	return nil
}

func fetchServices(c *cli.Context) error {
	// fetch all the index.yml files, and the service.yml files that they refer to
	// and put into a cache dir.
	// Q - should there be one dir per index.yml so that you can have more than one service with the same name..

	//TODO: need a --purge
	//TODO: also need to fetch the rancher/os choices

	return network.FetchAllServices()
}

func listServices(c *cli.Context) error {
	//get the current cfg, and the make a cfg with all cached services
	//then iterate through, listing all possible services, and what version is running, vs what version they could run
	//Can't just merge current cfg and cached services, as we lose service.yml version info

	currentConfig := config.LoadConfig()
	cachedConfigs := GetAllServices()
	// TODO: sort them!
	fmt.Printf("Enabled\n")
	enabledServices := make([]string, len(currentConfig.Rancher.Services)+len(currentConfig.Rancher.ServicesInclude))
	i := 0
	for k, _ := range currentConfig.Rancher.Services {
		enabledServices[i] = k
		i++
	}
	for k, _ := range currentConfig.Rancher.ServicesInclude {
		enabledServices[i] = k
		i++
	}
	sort.Strings(enabledServices)
	for _, serviceName := range enabledServices {
		// TODO: add running / stopped, error etc state
		// TODO: separate the volumes out too (they don't need the image listed - list the volumes instead)
		serviceConfig, _ := currentConfig.Rancher.Services[serviceName]
		if serviceConfig != nil {
			fmt.Printf("\t%s: %s\n", serviceName, serviceConfig.Image)

		} else {
			fmt.Printf("\t%s\n", serviceName)

		}
		if len(cachedConfigs[serviceName]) > 0 {
			fmt.Printf("\t\tAlternatives: ")
			for serviceLongName, _ := range cachedConfigs[serviceName] {
				fmt.Printf("%s, ", serviceLongName)
			}
			fmt.Printf("\n")
		}
	}
	fmt.Printf("Available\n")
	for serviceName, service := range cachedConfigs {
		if _, ok := currentConfig.Rancher.Services[serviceName]; ok {
			continue
		}
		for serviceLongName, _ := range service {
			fmt.Printf("\t%s: %s\n", serviceName, serviceLongName)
		}
	}

	return nil
}

func GetAllServices() map[string]map[string]*libcomposeConfig.ServiceConfigV1 {
	//result := make(map[string]*libcomposeConfig.ServiceConfig)
	result := make(map[string]map[string]*libcomposeConfig.ServiceConfigV1)

	cfg := config.LoadConfig()
	for repoName, _ := range cfg.Rancher.Repositories {
		indexPath := fmt.Sprintf("%s/index.yml", repoName)
		//content, err := network.LoadResource(indexPath, false)
		content, err := network.CacheLookup(indexPath)
		if err != nil {
			log.Errorf("Failed to load %s: %v", indexPath, err)
			continue
		}

		services := make(map[string][]string)
		err = yaml.Unmarshal(content, &services)
		if err != nil {
			log.Errorf("Failed to unmarshal %s: %v", indexPath, err)
			continue
		}
		for serviceType, serviceList := range services {
			for _, serviceLongName := range serviceList {
				p, err := service.LoadService(repoName, serviceLongName)
				if err != nil {
					log.Errorf("Failed to load %s/%s : %v", repoName, serviceLongName, err)
				}

				// yes, the serviceLongName is really only the yml file name
				// and each yml file can contain more than one actual service
				for serviceName, service := range p.Rancher.Services {
					//service, _ := p.ServiceConfigs.Get(serviceName)
					n := fmt.Sprintf("%s/%s", repoName, serviceLongName)
					if result[serviceName] == nil {
						result[serviceName] = map[string]*libcomposeConfig.ServiceConfigV1{}
					}
					result[serviceName][n] = service
					log.Debugf("loaded %s(%s): %s", n, serviceType, serviceName)
				}

			}
		}
	}

	return result
}

var originalCli = []cli.Command{
	{
		Name:        "config",
		ShortName:   "c",
		Usage:       "configure settings",
		HideHelp:    true,
		Subcommands: configSubcommands(),
	},
	{
		Name:        "console",
		Usage:       "manage which console container is used",
		HideHelp:    true,
		Subcommands: consoleSubcommands(),
	},
	{
		Name:        "engine",
		Usage:       "manage which Docker engine is used",
		HideHelp:    true,
		Subcommands: engineSubcommands(),
	},
	service.Commands(),
	{
		Name:        "os",
		Usage:       "operating system upgrade/downgrade",
		HideHelp:    true,
		Subcommands: osSubcommands(),
	},
	{
		Name:        "tls",
		Usage:       "setup tls configuration",
		HideHelp:    true,
		Subcommands: tlsConfCommands(),
	},
	installCommand,
	selinuxCommand(),
}

var hiddenInternalCommands = []cli.Command{
	{
		Name:            "bootstrap",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          bootstrapAction,
	},
	{
		Name:            "udev-settle",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          udevSettleAction,
	},
	{
		Name:            "user-docker",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          userDockerAction,
	},
	{
		Name:            "preload-images",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          preloadImagesAction,
	},
	{
		Name:            "switch-console",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          switchConsoleAction,
	},
	{
		Name:            "entrypoint",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          entrypointAction,
	},
	{
		Name:            "env",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          envAction,
	},
	{
		Name:            "console-init",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          consoleInitAction,
	},
	{
		Name:            "dev",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          devAction,
	},
	{
		Name:            "docker-init",
		Hidden:          true,
		HideHelp:        true,
		SkipFlagParsing: true,
		Action:          dockerInitAction,
	},
}
