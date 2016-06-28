package main

import (
	"github.com/containernetworking/cni/plugins/ipam/host-local"
	"github.com/containernetworking/cni/plugins/main/bridge"
	"github.com/docker/docker/docker"
	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/cniglue"
	"github.com/rancher/docker-from-scratch"
	"github.com/rancher/os/cmd/cloudinit"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/cmd/network"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/cmd/respawn"
	"github.com/rancher/os/cmd/switchconsole"
	"github.com/rancher/os/cmd/sysinit"
	"github.com/rancher/os/cmd/systemdocker"
	"github.com/rancher/os/cmd/userdocker"
	"github.com/rancher/os/cmd/wait"
	osInit "github.com/rancher/os/init"
)

var entrypoints = map[string]func(){
	"cloud-init":      cloudinit.Main,
	"docker":          docker.Main,
	"dockerlaunch":    dockerlaunch.Main,
	"halt":            power.Halt,
	"init":            osInit.MainInit,
	"netconf":         network.Main,
	"poweroff":        power.PowerOff,
	"reboot":          power.Reboot,
	"respawn":         respawn.Main,
	"ros-sysinit":     sysinit.Main,
	"shutdown":        power.Main,
	"switch-console":  switchconsole.Main,
	"system-docker":   systemdocker.Main,
	"user-docker":     userdocker.Main,
	"wait-for-docker": wait.Main,
	"cni-glue":        glue.Main,
	"bridge":          bridge.Main,
	"host-local":      hostlocal.Main,
}

func main() {
	for name, f := range entrypoints {
		reexec.Register(name, f)
	}

	if !reexec.Init() {
		control.Main()
	}
}
