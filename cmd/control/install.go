package control

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"

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
			Name:  "image, i",
			Usage: "install from a certain image",
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

	force := c.Bool("force")
	reboot := !c.Bool("no-reboot")

	if err := runInstall(image, installType, cloudConfig, device, force, reboot); err != nil {
		log.WithFields(log.Fields{"err": err}).Fatal("Failed to run install")
	}

	return nil
}

func runInstall(image, installType, cloudConfig, device string, force, reboot bool) error {
	in := bufio.NewReader(os.Stdin)

	fmt.Printf("Installing from %s\n", image)

	if !force {
		if !yes(in, "Continue") {
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
	cmd := exec.Command("system-docker", "run", "--net=host", "--privileged", "--volumes-from=user-volumes", image,
		"-d", device, "-t", installType, "-c", cloudConfig)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	if reboot && yes(in, "Continue with reboot") {
		log.Info("Rebooting")
		power.Reboot()
	}

	return nil
}
