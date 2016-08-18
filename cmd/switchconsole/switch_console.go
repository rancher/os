package switchconsole

import (
	"os"

	log "github.com/Sirupsen/logrus"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"golang.org/x/net/context"
)

func Main() {
	if len(os.Args) != 2 {
		log.Fatal("Must specify exactly one existing container")
	}
	newConsole := os.Args[1]

	cfg := config.LoadConfig()

	project, err := compose.GetProject(cfg, true, false)
	if err != nil {
		log.Fatal(err)
	}

	if newConsole != "default" {
                project.ServiceConfigs.Add("console", &composeConfig.ServiceConfig{})

		if err = compose.LoadService(project, cfg, true, newConsole); err != nil {
			log.Fatal(err)
		}
	}

	if err = config.Set("rancher.console", newConsole); err != nil {
		log.Errorf("Failed to update 'rancher.console': %v", err)
	}

	if err = project.Up(context.Background(), options.Up{
		Log: true,
	}, "console"); err != nil {
		log.Fatal(err)
	}

	if err = project.Restart(context.Background(), 10, "docker"); err != nil {
		log.Errorf("Failed to restart Docker: %v", err)
	}
}
