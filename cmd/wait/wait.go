package wait

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

func Main() {
	_, err := docker.NewClient(config.DOCKER_HOST)
	if err != nil {
		logrus.Errorf("Failed to conect to Docker")
		os.Exit(1)
	}

	logrus.Infof("Docker is ready")
}
