package user

import (
	"os"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

func SystemDocker() {
	var newEnv []string
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "DOCKER_HOST=") {
			newEnv = append(newEnv, env)
		}
	}

	newEnv = append(newEnv, "DOCKER_HOST=unix://var/run/system-docker.sock")

	os.Args[0] = "/usr/bin/docker"
	if err := syscall.Exec(os.Args[0], os.Args, newEnv); err != nil {
		log.Fatal(err)
	}
}
