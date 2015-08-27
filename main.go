package main

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/reexec"
	dockerlaunchMain "github.com/rancher/docker-from-scratch/main"
	"github.com/rancherio/os/cmd/cloudinit"
	"github.com/rancherio/os/cmd/control"
	"github.com/rancherio/os/cmd/network"
	"github.com/rancherio/os/cmd/power"
	"github.com/rancherio/os/cmd/respawn"
	"github.com/rancherio/os/cmd/sysinit"
	"github.com/rancherio/os/cmd/systemdocker"
	"github.com/rancherio/os/cmd/userdocker"
	"github.com/rancherio/os/cmd/wait"
	"github.com/rancherio/os/config"
	osInit "github.com/rancherio/os/init"
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
	registerCmd(config.SYSINIT_BIN, sysinit.Main)
	registerCmd("/usr/bin/dockerlaunch", dockerlaunchMain.Main)
	registerCmd("/usr/bin/user-docker", userdocker.Main)
	registerCmd("/usr/bin/system-docker", systemdocker.Main)
	registerCmd("/sbin/poweroff", power.PowerOff)
	registerCmd("/sbin/reboot", power.Reboot)
	registerCmd("/sbin/halt", power.Halt)
	registerCmd("/sbin/shutdown", power.Main)
	registerCmd("/usr/bin/respawn", respawn.Main)
	registerCmd("/usr/sbin/rancherctl", control.Main) // deprecated, use `ros` instead
	registerCmd("/usr/sbin/ros", control.Main)
	registerCmd("/usr/bin/cloud-init", cloudinit.Main)
	registerCmd("/usr/sbin/netconf", network.Main)
	registerCmd("/usr/sbin/wait-for-docker", wait.Main)

	if !reexec.Init() {
		log.Fatalf("Failed to find an entry point for %s", os.Args[0])
	}
}
