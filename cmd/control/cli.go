package control

import (
	"fmt"
	"os"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/codegangsta/cli"
	libcomposeConfig "github.com/docker/libcompose/config"

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
			Name:      "install",
			ShortName: "",
			Usage:     "install/upgrade service / RancherOS",
			HideHelp:  true,
			Action:    dummy,
		}, {
			Name:      "remove",
			ShortName: "",
			Usage:     "remove service",
			HideHelp:  true,
			Action:    dummy,
		}, {
			Name:      "logs",
			ShortName: "",
			Usage:     "service logs",
			HideHelp:  true,
			Action:    dummy,
		},
		// settings
		{
			Name:      "get",
			ShortName: "",
			Usage:     "get config value(s)",
			HideHelp:  true,
			Action:    dummy,
		}, {
			Name:      "set",
			ShortName: "",
			Usage:     "set config value(s)",
			HideHelp:  true,
			Action:    dummy,
		},
		// complete config
		{
			Name:      "export",
			ShortName: "",
			Usage:     "export config",
			HideHelp:  true,
			Action:    dummy,
		}, {
			Name:      "apply",
			ShortName: "",
			Usage:     "apply service&config changes",
			HideHelp:  true,
			Action:    dummy,
		}, {
			Name:      "validate",
			ShortName: "",
			Usage:     "validate config / service file",
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
	fmt.Printf("Running\n")
	for serviceName, serviceConfig := range currentConfig.Rancher.Services {
		fmt.Printf("\t%s: %s\n", serviceName, serviceConfig.Image)
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
				servicePath := fmt.Sprintf("%s/%s.yml", repoName, serviceLongName)
				//log.Infof("loading %s", serviceLongName)
				content, err := network.CacheLookup(servicePath)
				if err != nil {
					log.Errorf("Failed to load %s: %v", servicePath, err)
					continue
				}
				if content, err = ComposeToCloudConfig(content); err != nil {
					log.Errorf("Failed to convert compose to cloud-config syntax: %v", err)
					continue
				}

				p, err := config.ReadConfig(content, true)
				if err != nil {
					log.Errorf("Failed to load %s : %v", servicePath, err)
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

//TODO: copied from cloudinitsave, move to config.
func ComposeToCloudConfig(bytes []byte) ([]byte, error) {
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
