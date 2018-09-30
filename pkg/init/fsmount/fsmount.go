package fsmount

import (
	"fmt"
	"syscall"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/init/bootstrap"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
)

const (
	tmpfsMagic int64 = 0x01021994
	ramfsMagic int64 = 0x858458f6
)

var (
	ShouldSwitchRoot bool
)

func MountOem(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if err := mountConfigured("oem", cfg.Rancher.State.OemDev, cfg.Rancher.State.OemFsType, config.OemDir); err != nil {
		log.Debugf("Not mounting OEM: %v", err)
	} else {
		log.Infof("Mounted OEM: %s", cfg.Rancher.State.OemDev)
	}

	return cfg, nil
}

func MountBoot(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if IsInitrd() {
		return cfg, nil
	}

	if err := mountConfigured("boot", cfg.Rancher.State.BootDev, cfg.Rancher.State.BootFsType, config.BootDir); err != nil {
		log.Debugf("Not mounting BOOT: %v", err)
	} else {
		log.Infof("Mounted BOOT: %s", cfg.Rancher.State.BootDev)
	}

	return cfg, nil
}

func mountConfigured(display, dev, fsType, target string) error {
	var err error

	if dev == "" {
		return nil
	}

	dev = util.ResolveDevice(dev)
	if dev == "" {
		return fmt.Errorf("Could not resolve device %q", dev)
	}
	if fsType == "auto" {
		fsType, err = util.GetFsType(dev)
	}

	if err != nil {
		return err
	}

	log.Debugf("FsType has been set to %s", fsType)
	log.Infof("Mounting %s device %s to %s", display, dev, target)
	return util.Mount(dev, target, fsType, "")
}

func mountState(cfg *config.CloudConfig) error {
	return mountConfigured("state", cfg.Rancher.State.Dev, cfg.Rancher.State.FsType, config.StateDir)
}

func tryMountState(cfg *config.CloudConfig) error {
	err := mountState(cfg)
	if err == nil {
		return nil
	}
	log.Infof("Skipped an error when first mounting: %v", err)

	// If we failed to mount lets run bootstrap and try again
	if err := bootstrap.Bootstrap(cfg); err != nil {
		return err
	}

	return mountState(cfg)
}

func tryMountStateAndBootstrap(cfg *config.CloudConfig) (*config.CloudConfig, bool, error) {
	if !IsInitrd() || cfg.Rancher.State.Dev == "" {
		return cfg, false, nil
	}

	if err := tryMountState(cfg); !cfg.Rancher.State.Required && err != nil {
		return cfg, false, nil
	} else if err != nil {
		return cfg, false, err
	}

	return cfg, true, nil
}

func IsInitrd() bool {
	var stat syscall.Statfs_t
	syscall.Statfs("/", &stat)
	return int64(stat.Type) == tmpfsMagic || int64(stat.Type) == ramfsMagic
}

func MountStateAndBootstrap(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	var err error
	cfg, ShouldSwitchRoot, err = tryMountStateAndBootstrap(cfg)

	if err != nil {
		return nil, err
	}
	return cfg, nil
}
