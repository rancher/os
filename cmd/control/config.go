package control

import (
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

func configSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "get",
			Usage:  "get value",
			Action: configGet,
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

func configGet(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	content, err := cfg.Dump()
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[interface{}]interface{})
	yaml.Unmarshal([]byte(content), data)

	arg := c.Args().Get(0)
	if arg == "" {
		fmt.Println("")
		return
	}

	parts := strings.Split(arg, ".")
	for i, part := range parts {
		if val, ok := data[part]; ok {
			if i+1 == len(parts) {
				fmt.Println(val)
			} else {
				if newData, ok := val.(map[interface{}]interface{}); ok {
					data = newData
				} else {
					fmt.Println(val)
					break
				}
			}
		} else {
			fmt.Println("2")
			break
		}
	}
}

func configSave(c *cli.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	//TODO: why doesn't this work
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
