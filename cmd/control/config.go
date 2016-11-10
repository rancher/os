package control

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
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
			Name:     "generate",
			Usage:    "Generate a configuration file from a template",
			Action:   runGenerate,
			HideHelp: true,
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
					Usage: "Include the generated private keys",
				},
				cli.BoolFlag{
					Name:  "full, f",
					Usage: "Export full configuration, including internal and default settings",
				},
			},
			Action: export,
		},
		{
			Name:   "merge",
			Usage:  "merge configuration from stdin",
			Action: merge,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read",
				},
			},
		},
		{
			Name:   "validate",
			Usage:  "validate configuration from stdin",
			Action: validate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input, i",
					Usage: "File from which to read",
				},
			},
		},
	}
}

func imagesFromConfig(cfg *config.CloudConfig) []string {
	imagesMap := map[string]int{}

	for _, service := range cfg.Rancher.BootstrapContainers {
		imagesMap[service.Image] = 1
	}
	for _, service := range cfg.Rancher.Services {
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

func runImages(c *cli.Context) error {
	configFile := c.String("input")
	cfg, err := config.ReadConfig(nil, false, configFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "file": configFile}).Fatalf("Could not read config from file")
	}
	images := imagesFromConfig(cfg)
	fmt.Println(strings.Join(images, " "))
	return nil
}

func runGenerate(c *cli.Context) error {
	if err := genTpl(os.Stdin, os.Stdout); err != nil {
		log.Fatalf("Failed to generate config, err: '%s'", err)
	}
	return nil
}

func genTpl(in io.Reader, out io.Writer) error {
	bytes, err := ioutil.ReadAll(in)
	if err != nil {
		log.Fatal("Could not read from stdin")
	}
	tpl := template.Must(template.New("osconfig").Parse(string(bytes)))
	return tpl.Execute(out, env2map(os.Environ()))
}

func env2map(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, s := range env {
		d := strings.Split(s, "=")
		m[d[0]] = d[1]
	}
	return m
}

func configSet(c *cli.Context) error {
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	if key == "" {
		return nil
	}

	err := config.Set(key, value)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func configGet(c *cli.Context) error {
	arg := c.Args().Get(0)
	if arg == "" {
		return nil
	}

	val, err := config.Get(arg)
	if err != nil {
		log.WithFields(log.Fields{"key": arg, "val": val, "err": err}).Fatal("config get: failed to retrieve value")
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

	return nil
}

func merge(c *cli.Context) error {
	bytes, err := inputBytes(c)
	if err != nil {
		log.Fatal(err)
	}

	if err = config.Merge(bytes); err != nil {
		log.Fatal(err)
	}

	return nil
}

func export(c *cli.Context) error {
	content, err := config.Export(c.Bool("private"), c.Bool("full"))
	if err != nil {
		log.Fatal(err)
	}

	output := c.String("output")
	if output == "" {
		fmt.Println(content)
	} else {
		err := util.WriteFileAtomic(output, []byte(content), 0400)
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func validate(c *cli.Context) error {
	bytes, err := inputBytes(c)
	if err != nil {
		log.Fatal(err)
	}
	validationErrors, err := config.Validate(bytes)
	if err != nil {
		log.Fatal(err)
	}
	for _, validationError := range validationErrors.Errors() {
		log.Error(validationError)
	}
	return nil
}

func inputBytes(c *cli.Context) ([]byte, error) {
	input := os.Stdin
	inputFile := c.String("input")
	if inputFile != "" {
		var err error
		input, err = os.Open(inputFile)
		if err != nil {
			return nil, err
		}
		defer input.Close()
	}
	return ioutil.ReadAll(input)
}
