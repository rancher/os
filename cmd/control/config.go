package control

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
			Name:   "set",
			Usage:  "set a value",
			Action: configSet,
		},
		{
			Name:   "import",
			Usage:  "import configuration from standard in or a file",
			Action: configImport,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read",
				},
			},
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

func getConfigData() (map[interface{}]interface{}, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	content, err := cfg.Dump()
	if err != nil {
		log.Fatal(err)
	}

	data := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(content), data)

	return data, err
}

func configImport(c *cli.Context) {
	var input io.Reader
	var err error
	input = os.Stdin

	inputFile := c.String("input")
	if inputFile != "" {
		input, err = os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	bytes, err := ioutil.ReadAll(input)
	if err != nil {
		log.Fatal(err)
	}

	err = mergeConfig(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func mergeConfig(bytes []byte) error {
	var newConfig config.Config

	err := yaml.Unmarshal(bytes, &newConfig)
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	_, err = cfg.Merge(newConfig)
	if err != nil {
		return err
	}

	err = cfg.Save()
	if err != nil {
		return err
	}

	return err
}

func configSet(c *cli.Context) {
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	if key == "" {
		return
	}

	data, err := getConfigData()
	getOrSetVal(key, data, value)

	bytes, err := yaml.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	err = mergeConfig(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func configGet(c *cli.Context) {
	arg := c.Args().Get(0)
	if arg == "" {
		return
	}

	data, err := getConfigData()
	if err != nil {
		log.Fatal(err)
	}

	val := getOrSetVal(arg, data, nil)

	printYaml := false
	switch val.(type) {
	case []interface{}:
		printYaml = true
	case map[interface{}]interface{}:
		printYaml = true
	}

	if printYaml {
		bytes, err := yaml.Marshal(val)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(bytes))
	} else {
		fmt.Println(val)
	}
}

func getOrSetVal(args string, data map[interface{}]interface{}, value interface{}) interface{} {
	parts := strings.Split(args, ".")

	for i, part := range parts {
		val, ok := data[part]
		last := i+1 == len(parts)

		if last && value != nil {
			if s, ok := value.(string); ok {
				value = config.DummyMarshall(s)
			}

			data[part] = value
			return value
		}

		if !ok {
			break
		}

		if last {
			return val
		}

		newData, ok := val.(map[interface{}]interface{})
		if !ok {
			break
		}

		data = newData
	}

	return ""
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
