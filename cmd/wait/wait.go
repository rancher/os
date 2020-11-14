package wait

import (
	"os"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/docker"
	"github.com/burmilla/os/pkg/log"
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
