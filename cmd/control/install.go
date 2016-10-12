package control

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

var installCommand = cli.Command{
	Name:     "install",
	Usage:    "install RancherOS to disk",
	HideHelp: true,
	Action:   installAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			// TODO: need to validate ? -i rancher/os:v0.3.1 just sat there.
			Name: "image, i",
			Usage: `install from a certain image (e.g., 'rancher/os:v0.7.0')
							use 'ros os list' to see what versions are available.`,
		},
		cli.StringFlag{
			Name: "install-type, t",
			Usage: `generic:    (Default) Creates 1 ext4 partition and installs RancherOS
                        amazon-ebs: Installs RancherOS and sets up PV-GRUB`,
		},
		cli.StringFlag{
			Name:  "cloud-config, c",
			Usage: "cloud-config yml file - needed for SSH authorized keys",
		},
		cli.StringFlag{
			Name:  "device, d",
			Usage: "storage device",
		},
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "[ DANGEROUS! Data loss can happen ] partition/format without prompting",
		},
		cli.BoolFlag{
			Name:  "no-reboot",
			Usage: "do not reboot after install",
		},
		cli.StringFlag{
			Name:  "append, a",
			Usage: "append additional kernel parameters",
		},
	},
}

func installAction(c *cli.Context) error {
	if c.Args().Present() {
		log.Fatalf("invalid arguments %v", c.Args())
	}
	device := c.String("device")
	if device == "" {
		log.Fatal("Can not proceed without -d <dev> specified")
	}

	image := c.String("image")
	cfg := config.LoadConfig()
	if image == "" {
		image = cfg.Rancher.Upgrade.Image + ":" + config.VERSION + config.SUFFIX
	}

	installType := c.String("install-type")
	if installType == "" {
		log.Info("No install type specified...defaulting to generic")
		installType = "generic"
	}

	cloudConfig := c.String("cloud-config")
	if cloudConfig == "" {
		log.Warn("Cloud-config not provided: you might need to provide cloud-config on boot with ssh_authorized_keys")
	} else {
		uc := "/opt/user_config.yml"
		if err := util.FileCopy(cloudConfig, uc); err != nil {
			log.WithFields(log.Fields{"cloudConfig": cloudConfig}).Fatal("Failed to copy cloud-config")
		}
		cloudConfig = uc
	}

	append := strings.TrimSpace(c.String("append"))
	force := c.Bool("force")
	reboot := !c.Bool("no-reboot")

	if err := runInstall(image, installType, cloudConfig, device, append, force, reboot); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to run install")
	}

	return nil
}

func runInstall(image, installType, cloudConfig, device, append string, force, reboot bool) error {
	fmt.Printf("Installing from %s\n", image)

	if !force {
		if !yes("Continue") {
			os.Exit(1)
		}
	}

	if installType == "generic" {
		cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=all-volumes",
			"--entrypoint=/scripts/set-disk-partitions", image, device)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=user-volumes",
		"--volumes-from=command-volumes", image, "-d", device, "-t", installType, "-c", cloudConfig, "-a", append)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if reboot && (force || yes("Continue with reboot")) {
		log.Info("Rebooting")
		power.Reboot()
	}

	return nil
}
