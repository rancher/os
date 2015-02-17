package main

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/reexec"
	"github.com/rancherio/os/cmd/control"
	"github.com/rancherio/os/cmd/systemdocker"
	osInit "github.com/rancherio/os/init"
	"github.com/rancherio/os/power"
	"github.com/rancherio/os/respawn"
	"github.com/rancherio/os/sysinit"
	"github.com/rancherio/os/util"
)

func registerCmd(cmd string, mainFunc func()) {
	log.Debugf("Registering main %s", cmd)
	reexec.Register(cmd, mainFunc)

	parts := strings.Split(cmd, "/")
	if len(parts) == 0 {
		return
	}

	last := parts[len(parts)-1]

	log.Debugf("Registering main %s", last)
	reexec.Register(last, mainFunc)

	log.Debugf("Registering main %s", "./"+last)
	reexec.Register("./"+last, mainFunc)
}

func main() {
	registerCmd("/init", osInit.MainInit)
	registerCmd("/sbin/init-sys", sysinit.SysInit)
	registerCmd("/usr/bin/system-docker", systemdocker.Main)
	registerCmd("/sbin/poweroff", power.PowerOff)
	registerCmd("/sbin/reboot", power.Reboot)
	registerCmd("/sbin/halt", power.Halt)
	registerCmd("/usr/bin/respawn", respawn.Main)
	registerCmd("/usr/sbin/rancherctl", control.Main)
        registerCmd("/sbin/tlsconf", util.TLSConf)
	
	if !reexec.Init() {
		log.Fatalf("Failed to find an entry point for %s", os.Args[0])
	}
}
