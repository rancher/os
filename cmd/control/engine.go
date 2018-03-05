package control

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/net/context"

	"github.com/codegangsta/cli"
	"github.com/docker/docker/reference"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/cmd/control/service"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

func engineSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "switch",
			Usage:  "switch Docker engine without a reboot",
			Action: engineSwitch,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
				cli.BoolFlag{
					Name:  "no-pull",
					Usage: "don't pull console image",
				},
			},
		},
		{
			Name:   "enable",
			Usage:  "set Docker engine to be switched on next reboot",
			Action: engineEnable,
		},
		{
			Name:   "list",
			Usage:  "list available Docker engines",
			Action: engineList,
		},
	}
}

func engineSwitch(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to switch to")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()
	validateEngine(newEngine, cfg)

	project, err := compose.GetProject(cfg, true, false)
	if err != nil {
		log.Fatal(err)
	}

	if err = project.Stop(context.Background(), 10, "docker"); err != nil {
		log.Fatal(err)
	}

	if err = compose.LoadSpecialService(project, cfg, "docker", newEngine); err != nil {
		log.Fatal(err)
	}

	if err = project.Up(context.Background(), options.Up{}, "docker"); err != nil {
		log.Fatal(err)
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update rancher.docker.engine: %v", err)
	}

	return nil
}

func engineEnable(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to enable")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()
	validateEngine(newEngine, cfg)

	if err := compose.StageServices(cfg, newEngine); err != nil {
		return err
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update 'rancher.docker.engine': %v", err)
	}

	return nil
}

func engineList(c *cli.Context) error {
	cfg := config.LoadConfig()
	engines := availableEngines(cfg)
	currentEngine := CurrentEngine()

	for _, engine := range engines {
		if engine == currentEngine {
			fmt.Printf("current  %s\n", engine)
		} else if engine == cfg.Rancher.Docker.Engine {
			fmt.Printf("enabled  %s\n", engine)
		} else {
			fmt.Printf("disabled %s\n", engine)
		}
	}

	return nil
}

func validateEngine(engine string, cfg *config.CloudConfig) {
	engines := availableEngines(cfg)
	if !service.IsLocalOrURL(engine) && !util.Contains(engines, engine) {
		log.Fatalf("%s is not a valid engine", engine)
	}
}

func availableEngines(cfg *config.CloudConfig) []string {
	engines, err := network.GetEngines(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(engines)
	return engines
}

// CurrentEngine gets the name of the docker that's running
func CurrentEngine() (engine string) {
	// sudo system-docker inspect --format "{{.Config.Image}}" docker
	client, err := docker.NewSystemClient()
	if err != nil {
		log.Warnf("Failed to detect current docker: %v", err)
		return
	}
	info, err := client.ContainerInspect(context.Background(), "docker")
	if err != nil {
		log.Warnf("Failed to detect current docker: %v", err)
		return
	}
	// parse image name, then remove os- prefix and the engine suffix
	image, err := reference.ParseNamed(info.Config.Image)
	if err != nil {
		log.Warnf("Failed to detect current docker(%s): %v", info.Config.Image, err)
		return
	}
	if t, ok := image.(reference.NamedTagged); ok {
		tag := t.Tag()
		if !strings.HasPrefix(tag, "1.") {
			// TODO: this assumes we only do Docker ce :/
			tag = tag + "-ce"
		}
		return "docker-" + tag
	}

	return
}
