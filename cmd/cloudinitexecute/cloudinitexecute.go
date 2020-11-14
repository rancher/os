package cloudinitexecute

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	rancherConfig "github.com/burmilla/os/config"
	"github.com/burmilla/os/config/cloudinit/config"
	"github.com/burmilla/os/config/cloudinit/system"
	"github.com/burmilla/os/pkg/docker"
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"

	"golang.org/x/net/context"
)

const (
	resizeStamp = "/var/lib/rancher/resizefs.done"
	sshKeyName  = "rancheros-cloud-config"
)

var (
	console    bool
	preConsole bool
	flags      *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&console, "console", false, "apply console configuration")
	flags.BoolVar(&preConsole, "pre-console", false, "apply pre-console configuration")
}

func Main() {
	flags.Parse(os.Args[1:])

	log.InitLogger()
	log.Infof("Running cloud-init-execute: pre-console=%v, console=%v", preConsole, console)

	cfg := rancherConfig.LoadConfig()

	if !console && !preConsole {
		console = true
		preConsole = true
	}

	if console {
		ApplyConsole(cfg)
	}
	if preConsole {
		applyPreConsole(cfg)
	}
}

func ApplyConsole(cfg *rancherConfig.CloudConfig) {
	if len(cfg.SSHAuthorizedKeys) > 0 {
		if err := authorizeSSHKeys("rancher", cfg.SSHAuthorizedKeys, sshKeyName); err != nil {
			log.Error(err)
		}
		if err := authorizeSSHKeys("docker", cfg.SSHAuthorizedKeys, sshKeyName); err != nil {
			log.Error(err)
		}
	}

	WriteFiles(cfg, "console")

	for _, mount := range cfg.Mounts {
		if len(mount) != 4 {
			log.Errorf("Unable to mount %s: must specify exactly four arguments", mount[1])
		}

		if mount[2] == "nfs" || mount[2] == "nfs4" {
			if err := os.MkdirAll(mount[1], 0755); err != nil {
				log.Errorf("Unable to create mount point %s: %v", mount[1], err)
				continue
			}
			cmdArgs := []string{mount[0], mount[1], "-t", mount[2]}
			if mount[3] != "" {
				cmdArgs = append(cmdArgs, "-o", mount[3])
			}
			cmd := exec.Command("mount", cmdArgs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Errorf("Failed to mount %s: %v", mount[1], err)
			}
			continue
		}

		device := util.ResolveDevice(mount[0])

		if mount[2] == "swap" {
			cmd := exec.Command("swapon", device)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Errorf("Unable to swapon %s: %v", device, err)
			}
			continue
		}

		if err := util.Mount(device, mount[1], mount[2], mount[3]); err != nil {
			log.Errorf("Failed to mount %s: %v", mount[1], err)
		}
	}

	err := util.RunCommandSequence(cfg.Runcmd)
	if err != nil {
		log.Error(err)
	}
}

func WriteFiles(cfg *rancherConfig.CloudConfig, container string) {
	for _, file := range cfg.WriteFiles {
		fileContainer := file.Container
		if fileContainer == "" {
			fileContainer = "console"
		}
		if fileContainer != container {
			continue
		}

		content, err := config.DecodeContent(file.File.Content, file.File.Encoding)
		if err != nil {
			continue
		}

		file.File.Content = string(content)
		file.File.Encoding = ""

		f := system.File{
			File: file.File,
		}
		fullPath, err := system.WriteFile(&f, "/")
		if err != nil {
			log.WithFields(log.Fields{"err": err, "path": fullPath}).Error("Error writing file")
			continue
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}
}

func applyPreConsole(cfg *rancherConfig.CloudConfig) {
	if cfg.Rancher.ResizeDevice != "" {
		if _, err := os.Stat(resizeStamp); os.IsNotExist(err) {
			if err := resizeDevice(cfg); err == nil {
				os.Create(resizeStamp)
			} else {
				log.Errorf("Failed to resize %s: %s", cfg.Rancher.ResizeDevice, err)
			}
		} else {
			log.Infof("Skipped resizing %s because %s exists", cfg.Rancher.ResizeDevice, resizeStamp)
		}
	}

	for k, v := range cfg.Rancher.Sysctl {
		elems := []string{"/proc", "sys"}
		elems = append(elems, strings.Split(k, ".")...)
		path := path.Join(elems...)
		if err := ioutil.WriteFile(path, []byte(v), 0644); err != nil {
			log.Errorf("Failed to set sysctl key %s: %s", k, err)
		}
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		log.Error(err)
	}

	for _, restart := range cfg.Rancher.RestartServices {
		if err = client.ContainerRestart(context.Background(), restart, 10); err != nil {
			log.Error(err)
		}
	}
}

func resizeDevice(cfg *rancherConfig.CloudConfig) error {
	partition := "1"
	targetPartition := fmt.Sprintf("%s%s", cfg.Rancher.ResizeDevice, partition)

	if strings.Contains(cfg.Rancher.ResizeDevice, "mmcblk") {
		partition = "2"
		targetPartition = fmt.Sprintf("%sp%s", cfg.Rancher.ResizeDevice, partition)
	} else if strings.Contains(cfg.Rancher.ResizeDevice, "nvme") {
		targetPartition = fmt.Sprintf("%sp%s", cfg.Rancher.ResizeDevice, partition)
	}

	cmd := exec.Command("growpart", cfg.Rancher.ResizeDevice, partition)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()

	cmd = exec.Command("partprobe", cfg.Rancher.ResizeDevice)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("resize2fs", targetPartition)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
