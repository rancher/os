package control

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"golang.org/x/net/context"
)

func switchConsoleAction(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("Must specify exactly one existing container")
	}
	newConsole := c.Args()[0]

	cfg := config.LoadConfig()

	project, err := compose.GetProject(cfg, true, false)
	if err != nil {
		return err
	}

	if newConsole != "default" {
		if err = compose.LoadSpecialService(project, cfg, "console", newConsole); err != nil {
			return err
		}
	}

	if err = config.Set("rancher.console", newConsole); err != nil {
		log.Errorf("Failed to update 'rancher.console': %v", err)
	}

	if err = project.Up(context.Background(), options.Up{
		Log: true,
	}, "console"); err != nil {
		return err
	}

	if err = project.Restart(context.Background(), 10, "docker"); err != nil {
		log.Errorf("Failed to restart Docker: %v", err)
	}

	return nil
}
