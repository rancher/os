package control

import (
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

func configSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "get",
			Usage: "get value",
		},
		{
			Name:  "import",
			Usage: "list values",
		},
		{
			Name:  "export",
			Usage: "dump full configuration",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "File to which to save",
				},
				cli.BoolFlag{
					Name:  "full",
					Usage: "Include full configuration, not just writable fields",
				},
			},
			Action: configSave,
		},
	}
}

func configSave(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range cfg.SystemContainers {
		container := docker.NewContainer("", &c)
		if container.Err != nil {
			log.Fatalf("Failed to parse [%s] : %v", c.Cmd, container.Err)
		}
	}

	if !c.Bool("full") {
		cfg.ClearReadOnly()
	}

	content, err := cfg.Dump()

	output := c.String("output")
	if output == "" {
		fmt.Println(content)
	} else {
		err := ioutil.WriteFile(output, []byte(content), 0400)
		if err != nil {
			log.Fatal(err)
		}
	}

}
