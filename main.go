package main

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/docker-from-scratch"
	"github.com/rancher/os/cmd/cloudinit"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/cmd/network"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/cmd/respawn"
	"github.com/rancher/os/cmd/sysinit"
	"github.com/rancher/os/cmd/systemdocker"
	"github.com/rancher/os/cmd/userdocker"
	"github.com/rancher/os/cmd/wait"
	"github.com/rancher/os/config"
	osInit "github.com/rancher/os/init"
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
	registerCmd("/usr/bin/dockerlaunch", dockerlaunch.Main)
	registerCmd("/usr/bin/user-docker", userdocker.Main)
	registerCmd("/usr/bin/system-docker", systemdocker.Main)
	registerCmd("/sbin/poweroff", power.PowerOff)
	registerCmd("/sbin/reboot", power.Reboot)
	registerCmd("/sbin/halt", power.Halt)
	registerCmd("/sbin/shutdown", power.Main)
	registerCmd("/usr/bin/respawn", respawn.Main)
	registerCmd("/usr/bin/ros", control.Main)
	registerCmd("/usr/bin/cloud-init", cloudinit.Main)
	registerCmd("/usr/sbin/netconf", network.Main)
	registerCmd("/usr/sbin/wait-for-docker", wait.Main)

	if !reexec.Init() {
		reexec.Register(os.Args[0], control.Main)
		if !reexec.Init() {
			log.Fatalf("Failed to find an entry point for %s", os.Args[0])
		}
	}
}
