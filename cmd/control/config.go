package control

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
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
			Action: runImport,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read",
				},
			},
		},
		{
			Name:   "images",
			Usage:  "List Docker images for a configuration from a file",
			Action: runImages,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read config",
				},
			},
		},
		{
			Name:  "export",
			Usage: "export configuration",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "output, o",
					Usage: "File to which to save",
				},
				cli.BoolFlag{
					Name:  "private, p",
					Usage: "Include private information such as keys",
				},
				cli.BoolFlag{
					Name:  "full, f",
					Usage: "Include full configuration, including internal and default settings",
				},
			},
			Action: export,
		},
		{
			Name:   "merge",
			Usage:  "merge configuration from stdin",
			Action: merge,
		},
	}
}

func imagesFromConfig(cfg *config.Config) []string {
	imagesMap := map[string]int{}

	for _, service := range cfg.BootstrapContainers {
		imagesMap[service.Image] = 1
	}
	for _, service := range cfg.SystemContainers {
		imagesMap[service.Image] = 1
	}

	images := make([]string, len(imagesMap))
	i := 0
	for image := range imagesMap {
		images[i] = image
		i += 1
	}
	sort.Strings(images)
	return images
}

func runImages(c *cli.Context) {
	configFile := c.String("input")
	cfg := config.ReadConfig(configFile)
	if cfg == nil {
		log.Fatalf("Could not read config from file %v", configFile)
	}
	images := imagesFromConfig(cfg)
	fmt.Println(strings.Join(images, " "))
}

func runImport(c *cli.Context) {
	var input io.ReadCloser
	var err error
	input = os.Stdin
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal(err)
	}

	inputFile := c.String("input")
	if inputFile != "" {
		input, err = os.Open(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer input.Close()
	}

	bytes, err := ioutil.ReadAll(input)
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Import(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func configSet(c *cli.Context) {
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	if key == "" {
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Set(key, value)
	if err != nil {
		log.Fatal(err)
	}
}

func configGet(c *cli.Context) {
	arg := c.Args().Get(0)
	if arg == "" {
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	val, err := cfg.Get(arg)
	if err != nil {
		log.Fatal(err)
	}

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

		// Reached end, set the value
		if last && value != nil {
			if s, ok := value.(string); ok {
				value = config.DummyMarshall(s)
			}

			data[part] = value
			return value
		}

		// Missing intermediate key, create key
		if !last && value != nil && !ok {
			newData := map[interface{}]interface{}{}
			data[part] = newData
			data = newData
			continue
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

func merge(c *cli.Context) {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = cfg.Merge(bytes)
	if err != nil {
		log.Fatal(err)
	}
}

func export(c *cli.Context) {
	content, err := config.Dump(c.Bool("private"), c.Bool("full"))
	if err != nil {
		log.Fatal(err)
	}

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
