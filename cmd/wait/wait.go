package wait

import (
	"os"

	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"
)

func Main() {
	log.InitLogger()
	_, err := docker.NewClient(config.DockerHost)
	if err != nil {
		log.Errorf("Failed to connect to Docker")
		os.Exit(1)
	}

	log.Infof("Docker is ready")
}
