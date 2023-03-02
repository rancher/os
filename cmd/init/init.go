//go:build linux
// +build linux

package init

import (
	"fmt"

	"github.com/burmilla/os/config"
	"github.com/burmilla/os/pkg/dfs"
	"github.com/burmilla/os/pkg/init/b2d"
	"github.com/burmilla/os/pkg/init/cloudinit"
	"github.com/burmilla/os/pkg/init/configfiles"
	"github.com/burmilla/os/pkg/init/debug"
	"github.com/burmilla/os/pkg/init/docker"
	"github.com/burmilla/os/pkg/init/env"
	"github.com/burmilla/os/pkg/init/fsmount"
	"github.com/burmilla/os/pkg/init/hypervisor"
	"github.com/burmilla/os/pkg/init/modules"
	"github.com/burmilla/os/pkg/init/one"
	"github.com/burmilla/os/pkg/init/prepare"
	"github.com/burmilla/os/pkg/init/recovery"
	"github.com/burmilla/os/pkg/init/sharedroot"
	"github.com/burmilla/os/pkg/init/switchroot"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/sysinit"
)

func MainInit() {
	log.InitLogger()
	// TODO: this breaks and does nothing if the cfg is invalid (or is it due to threading?)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Starting Recovery console: %v\n", r)
			recovery.Recovery(nil)
		}
	}()

	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func RunInit() error {
	initFuncs := config.CfgFuncs{
		{Name: "set env", Func: env.Init},
		{Name: "preparefs", Func: prepare.FS},
		{Name: "save init cmdline", Func: prepare.SaveCmdline},
		{Name: "mount OEM", Func: fsmount.MountOem},
		{Name: "debug save cfg", Func: debug.PrintAndLoadConfig},
		{Name: "load modules", Func: modules.LoadModules},
		{Name: "recovery console", Func: recovery.LoadRecoveryConsole},
		{Name: "b2d env", Func: b2d.B2D},
		{Name: "mount STATE and bootstrap", Func: fsmount.MountStateAndBootstrap},
		{Name: "cloud-init", Func: cloudinit.CloudInit},
		{Name: "read cfg and log files", Func: configfiles.ReadConfigFiles},
		{Name: "switchroot", Func: switchroot.SwitchRoot},
		{Name: "mount OEM2", Func: fsmount.MountOem},
		{Name: "mount BOOT", Func: fsmount.MountBoot},
		{Name: "write cfg and log files", Func: configfiles.WriteConfigFiles},
		{Name: "b2d Env", Func: b2d.Env},
		{Name: "hypervisor tools", Func: hypervisor.Tools},
		{Name: "preparefs2", Func: prepare.FS},
		{Name: "load modules2", Func: modules.LoadModules},
		{Name: "set proxy env", Func: env.Proxy},
		{Name: "setupSharedRoot", Func: sharedroot.Setup},
		{Name: "sysinit", Func: sysinit.RunSysInit},
	}

	cfg, err := config.ChainCfgFuncs(nil, initFuncs)
	if err != nil {
		recovery.Recovery(err)
	}

	launchConfig, args := docker.GetLaunchConfig(cfg, &cfg.Rancher.SystemDocker)
	launchConfig.Fork = !cfg.Rancher.SystemDocker.Exec
	//launchConfig.NoLog = true

	log.Info("Launching System Docker")
	_, err = dfs.LaunchDocker(launchConfig, config.SystemDockerBin, args...)
	if err != nil {
		log.Errorf("Error Launching System Docker: %s", err)
		recovery.Recovery(err)
		return err
	}
	// Code never gets here - rancher.system_docker.exec=true

	return one.PidOne()
}
