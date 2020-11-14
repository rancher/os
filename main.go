package main

import (
	"fmt"
	"os"

	"github.com/burmilla/os/cmd/cloudinitexecute"
	"github.com/burmilla/os/cmd/cloudinitsave"
	"github.com/burmilla/os/cmd/control"
	osInit "github.com/burmilla/os/cmd/init"
	"github.com/burmilla/os/cmd/network"
	"github.com/burmilla/os/cmd/power"
	"github.com/burmilla/os/cmd/respawn"
	"github.com/burmilla/os/cmd/sysinit"
	"github.com/burmilla/os/cmd/wait"
	"github.com/burmilla/os/pkg/dfs"

	"github.com/docker/docker/pkg/reexec"
)

var entrypoints = map[string]func(){
	"autologin":          control.AutologinMain,
	"cloud-init-execute": cloudinitexecute.Main,
	"cloud-init-save":    cloudinitsave.Main,
	"dockerlaunch":       dfs.Main,
	"init":               osInit.MainInit,
	"netconf":            network.Main,
	"recovery":           control.AutologinMain,
	"ros-bootstrap":      control.BootstrapMain,
	"ros-sysinit":        sysinit.Main,
	"wait-for-docker":    wait.Main,
	"respawn":            respawn.Main,

	// Power commands
	"halt":     power.Shutdown,
	"poweroff": power.Shutdown,
	"reboot":   power.Shutdown,
	"shutdown": power.Shutdown,
}

func main() {
	if 0 == 1 {
		// TODO: move this into a "dev/debug +build"
		fmt.Fprintf(os.Stderr, "ros main(%s) ppid:%d - print to stdio\n", os.Args[0], os.Getppid())

		filename := "/dev/kmsg"
		f, err := os.OpenFile(filename, os.O_WRONLY, 0644)
		if err == nil {
			fmt.Fprintf(f, "ros main(%s) ppid:%d - print to %s\n", os.Args[0], os.Getppid(), filename)
		}
		f.Close()
	}

	for name, f := range entrypoints {
		reexec.Register(name, f)
	}

	if !reexec.Init() {
		control.Main()
	}
}
