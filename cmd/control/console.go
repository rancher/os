package control

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
	"github.com/rancher/os/util/network"
)

func consoleSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "switch",
			Usage:  "switch console without a reboot",
			Action: consoleSwitch,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
			},
		},
		{
			Name:   "enable",
			Usage:  "set console to be switched on next reboot",
			Action: consoleEnable,
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
		log.Fatal("Must specify exactly one console to switch to")
	}
	newConsole := c.Args()[0]

	cfg := config.LoadConfig()
	consoles, err := network.GetConsoles(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		return err
	}

	consoles = append(consoles, "default")
	if !util.Contains(consoles, newConsole) {
		log.Fatalf("Console %s is not a valid console, user\"sudo ros console list\" to list consoles", newConsole)
	}

	if newConsole == currentConsole() {
		log.Warnf("Console is already set to %s", newConsole)
	}

	if !c.Bool("force") {
		in := bufio.NewReader(os.Stdin)
		fmt.Println("Switching consoles will destroy the current console container and restart Docker.")
		fmt.Println("Note: You will also be logged out.")
		if !yes(in, "Continue") {
			return nil
		}
	}

	if newConsole != "default" {
		if err := compose.StageServices(cfg, newConsole); err != nil {
			return err
		}
	}

	service, err := compose.CreateService(nil, "switch-console", &composeConfig.ServiceConfigV1{
		LogDriver:  "json-file",
		Privileged: true,
		Net:        "host",
		Pid:        "host",
		Image:      config.OS_BASE,
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
	if err = service.Up(context.Background(), options.Up{}); err != nil {
		return err
	}
	return service.Log(context.Background(), true)
}

func consoleEnable(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one console to enable")
	}
	newConsole := c.Args()[0]

	cfg := config.LoadConfig()
	consoles, err := network.GetConsoles(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		return err
	}

	consoles = append(consoles, "default")
	if !util.Contains(consoles, newConsole) {
		log.Fatalf("Console %s is not a valid console, user\"sudo ros console list\" to list consoles", newConsole)
	}

	if newConsole != "default" {
		if err := compose.StageServices(cfg, newConsole); err != nil {
			return err
		}
	}

	if err := config.Set("rancher.console", newConsole); err != nil {
		log.Errorf("Failed to update 'rancher.console': %v", err)
	}

	return nil
}

func consoleList(c *cli.Context) error {
	cfg := config.LoadConfig()

	consoles, err := network.GetConsoles(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		return err
	}
	consoles = append(consoles, "default")
	sort.Strings(consoles)

	currentConsole := currentConsole()

	for _, console := range consoles {
		if console == currentConsole {
			fmt.Printf("current  %s\n", console)
		} else if console == cfg.Rancher.Console {
			fmt.Printf("enabled  %s\n", console)
		} else {
			fmt.Printf("disabled %s\n", console)
		}
	}

	return nil
}

func currentConsole() (console string) {
	consoleBytes, err := ioutil.ReadFile("/run/console-done")
	if err == nil {
		console = strings.TrimSpace(string(consoleBytes))
	} else {
		log.Warnf("Failed to detect current console: %v", err)
	}
	return
}
