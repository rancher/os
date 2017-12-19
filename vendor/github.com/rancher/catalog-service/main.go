package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/rancher/catalog-service/cmd"
	_ "github.com/rancher/catalog-service/signals"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
