package control

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

func consoleSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "switch",
			Usage:  "switch currently running console",
			Action: consoleSwitch,
		},
		{
			Name:   "list",
			Usage:  "list available consoles",
			Action: consoleList,
		},
	}
}

func consoleSwitch(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one existing container")
	}
	newConsole := c.Args()[0]

	in := bufio.NewReader(os.Stdin)
	question := fmt.Sprintf("Switching consoles will destroy the current console container and restart Docker. Continue")
	if !yes(in, question) {
		return nil
	}

	cfg := config.LoadConfig()

	if err := compose.StageServices(cfg, newConsole); err != nil {
		return err
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}

	currentContainerId, err := util.GetCurrentContainerId()
	if err != nil {
		return err
	}

	currentContainer, err := client.ContainerInspect(context.Background(), currentContainerId)
	if err != nil {
		return err
	}

	service, err := compose.CreateService(nil, "switch-console", &composeConfig.ServiceConfigV1{
		LogDriver:  "json-file",
		Privileged: true,
		Net:        "host",
		Pid:        "host",
		Image:      currentContainer.Config.Image,
		Labels: map[string]string{
			config.SCOPE: config.SYSTEM,
		},
		Command:     []string{"/usr/bin/switch-console", newConsole},
		VolumesFrom: []string{"all-volumes"},
	})
	if err != nil {
		return err
	}

	if err = service.Delete(context.Background(), options.Delete{}); err != nil {
		return err
	}
	return service.Up(context.Background(), options.Up{})
}

func consoleList(c *cli.Context) error {
	cfg := config.LoadConfig()

	consoles, err := network.GetConsoles(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		return err
	}

	for _, console := range consoles {
		fmt.Println(console)
	}

	return nil
}
