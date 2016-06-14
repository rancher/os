package systemdocker

import (
	"log"
	"os"

	"github.com/docker/docker/docker"
	"github.com/rancher/os/config"
)

func Main() {
	if os.Geteuid() != 0 {
		log.Fatalf("%s: Need to be root", os.Args[0])
	}

	if os.Getenv("DOCKER_HOST") == "" {
		os.Setenv("DOCKER_HOST", config.DOCKER_SYSTEM_HOST)
	}

	docker.Main()
}
