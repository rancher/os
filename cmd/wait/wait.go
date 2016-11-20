package wait

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
)

func Main() {
	_, err := docker.NewClient(config.DOCKER_HOST)
	if err != nil {
		logrus.Errorf("Failed to connect to Docker")
		os.Exit(1)
	}

	logrus.Infof("Docker is ready")
}
