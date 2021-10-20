package main

import (
	"flag"
	"os"

	"github.com/rancher/os2/pkg/config"
	"github.com/rancher/os2/pkg/install"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

var (
	output      = flag.Bool("automatic", false, "Check for and run automatic installation")
	printConfig = flag.Bool("print-config", false, "Print effective configuration and exit")
	configFile  = flag.String("config-file", "", "Config file to use, local file or http/tftp URL")
)

func main() {
	flag.Parse()
	if *printConfig {
		cfg, err := config.ReadConfig(*configFile)
		if err != nil {
			logrus.Fatal(err)
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
		os.Stdout.Write(data)
		return
	}

	if err := install.Run(*output, *configFile); err != nil {
		logrus.Fatal(err)
	}
}
