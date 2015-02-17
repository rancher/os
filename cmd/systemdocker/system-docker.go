package systemdocker

import (
	"os"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/config"
)

func Main() {
	var newEnv []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "DOCKER_HOST=") {
			newEnv = append(newEnv, env)
		}
	}

	newEnv = append(newEnv, "DOCKER_HOST="+config.DOCKER_SYSTEM_HOST)

	os.Args[0] = "/usr/bin/docker"
	if err := syscall.Exec(os.Args[0], os.Args, newEnv); err != nil {
		log.Fatal(err)
	}
}
