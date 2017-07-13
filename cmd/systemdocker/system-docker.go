package systemdocker

import (
	"os"

	"github.com/docker/docker/docker"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func Main() {
	log.SetLevel(log.DebugLevel)

	if os.Geteuid() != 0 {
		log.Fatalf("%s: Need to be root", os.Args[0])
	}

	if os.Getenv("DOCKER_HOST") == "" {
		os.Setenv("DOCKER_HOST", config.SystemDockerHost)
	}

	docker.RancherOSMain()
}
