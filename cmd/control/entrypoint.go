package control

import (
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"golang.org/x/net/context"

	"github.com/rancher/os/cmd/cloudinitexecute"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/util"
)

const (
	ca     = "/etc/ssl/certs/ca-certificates.crt"
	caBase = "/etc/ssl/certs/ca-certificates.crt.rancher"
)

func entrypointAction(c *cli.Context) error {
	if _, err := os.Stat("/host/dev"); err == nil {
		cmd := exec.Command("mount", "--rbind", "/host/dev", "/dev")
		if err := cmd.Run(); err != nil {
			log.Errorf("Failed to mount /dev: %v", err)
		}
	}

	if err := util.FileCopy(caBase, ca); err != nil && !os.IsNotExist(err) {
		log.Error(err)
	}

	cfg := config.LoadConfig()

	shouldWriteFiles := false
	for _, file := range cfg.WriteFiles {
		if file.Container != "" {
			shouldWriteFiles = true
		}
	}

	if shouldWriteFiles {
		writeFiles(cfg)
	}

	setupPowerOperations()

	if len(os.Args) < 3 {
		return nil
	}

	binary, err := exec.LookPath(os.Args[2])
	if err != nil {
		return err
	}

	return syscall.Exec(binary, os.Args[2:], os.Environ())
}

func writeFiles(cfg *config.CloudConfig) error {
	id, err := util.GetCurrentContainerId()
	if err != nil {
		return err
	}
	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}
	info, err := client.ContainerInspect(context.Background(), id)
	if err != nil {
		return err
	}

	cloudinitexecute.WriteFiles(cfg, info.Name[1:])
	return nil
}

func setupPowerOperations() {
	for _, powerOperation := range []string{
		"/sbin/poweroff",
		"/sbin/shutdown",
		"/sbin/reboot",
		"/sbin/halt",
		"/usr/sbin/poweroff",
		"/usr/sbin/shutdown",
		"/usr/sbin/reboot",
		"/usr/sbin/halt",
	} {
		os.Remove(powerOperation)
	}

	for _, link := range []symlink{
		{config.ROS_BIN, "/sbin/poweroff"},
		{config.ROS_BIN, "/sbin/reboot"},
		{config.ROS_BIN, "/sbin/halt"},
		{config.ROS_BIN, "/sbin/shutdown"},
	} {
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}
}
