package main

import (
	"github.com/docker/docker/docker"
	"github.com/docker/docker/pkg/reexec"
)

func main() {
	if reexec.Init() {
		return
	}

	docker.Main()
}
