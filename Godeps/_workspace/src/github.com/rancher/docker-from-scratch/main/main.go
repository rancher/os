package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	dockerlaunch "github.com/rancher/docker-from-scratch"
)

func main() {
	Main()
}

func Main() {
	if os.Getenv("DOCKER_LAUNCH_DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage Example: %s /usr/bin/docker -d -D", os.Args[0])
	}

	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[2:]
	}

	var config dockerlaunch.Config
	args = dockerlaunch.ParseConfig(&config, args...)

	log.Debugf("Launch config %#v", config)

	_, err := dockerlaunch.LaunchDocker(&config, os.Args[1], args...)
	if err != nil {
		log.Fatal(err)
	}
}
