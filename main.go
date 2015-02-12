package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/reexec"
	osInit "github.com/rancherio/os/init"
	"github.com/rancherio/os/power"
	"github.com/rancherio/os/respawn"
	"github.com/rancherio/os/sysinit"
	"github.com/rancherio/os/user"
)

func main() {
	reexec.Register("init", osInit.MainInit)
	reexec.Register("/init", osInit.MainInit)
	reexec.Register("./init", osInit.MainInit)
	reexec.Register("/sbin/init-sys", sysinit.SysInit)
	reexec.Register("/usr/bin/system-docker", user.SystemDocker)
	reexec.Register("system-docker", user.SystemDocker)
	reexec.Register("poweroff", power.PowerOff)
	reexec.Register("reboot", power.Reboot)
	reexec.Register("halt", power.Halt)
	reexec.Register("respawn", respawn.Main)
	if !reexec.Init() {
		log.Fatalf("Failed to find an entry point for %s", os.Args[0])
	}
}
