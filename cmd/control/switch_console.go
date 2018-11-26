package control

import (
	"errors"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/compose"
	"github.com/rancher/os/pkg/log"

	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project/options"
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

	// stop docker and console to avoid zombie process
	if err = project.Stop(context.Background(), 10, "docker"); err != nil {
		log.Errorf("Failed to stop Docker: %v", err)
	}
	if err = project.Stop(context.Background(), 10, "console"); err != nil {
		log.Errorf("Failed to stop console: %v", err)
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

	if err = project.Start(context.Background(), "docker"); err != nil {
		log.Errorf("Failed to start Docker: %v", err)
	}

	return nil
}
