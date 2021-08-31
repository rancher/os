package main

import (
	"flag"

	"github.com/rancher/os/pkg/install"
	"github.com/sirupsen/logrus"
)

var (
	output = flag.Bool("automatic", false, "Check for and run automatic installation")
)

func main() {
	flag.Parse()
	if err := install.Run(*output); err != nil {
		logrus.Fatal(err)
	}
}
