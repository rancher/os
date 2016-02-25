package control

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	dockerClient "github.com/fsouza/go-dockerclient"

	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/cmd/power"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
)

type Images struct {
	Current   string   `yaml:"current,omitempty"`
	Available []string `yaml:"available,omitempty"`
}

func osSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "upgrade",
			Usage:  "upgrade to latest version",
			Action: osUpgrade,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "stage, s",
					Usage: "Only stage the new upgrade, don't apply it",
				},
				cli.StringFlag{
					Name:  "image, i",
					Usage: "upgrade to a certain image",
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
				cli.BoolFlag{
					Name:  "no-reboot",
					Usage: "do not reboot after upgrade",
				},
				cli.BoolFlag{
					Name:  "kexec",
					Usage: "reboot using kexec",
				},
			},
		},
		{
			Name:   "list",
			Usage:  "list the current available versions",
			Action: osMetaDataGet,
		},
		{
			Name:   "version",
			Usage:  "show the currently installed version",
			Action: osVersion,
		},
	}
}

func getImages() (*Images, error) {
	upgradeUrl, err := getUpgradeUrl()
	if err != nil {
		return nil, err
	}

	var body []byte

	if strings.HasPrefix(upgradeUrl, "/") {
		body, err = ioutil.ReadFile(upgradeUrl)
		if err != nil {
			return nil, err
		}
	} else {
		u, err := url.Parse(upgradeUrl)
		if err != nil {
			return nil, err
		}

		q := u.Query()
		q.Set("current", config.VERSION)
		u.RawQuery = q.Encode()
		upgradeUrl = u.String()

		resp, err := http.Get(upgradeUrl)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	return parseBody(body)
}

func osMetaDataGet(c *cli.Context) {
	images, err := getImages()
	if err != nil {
		log.Fatal(err)
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		log.Fatal(err)
	}

	for _, image := range images.Available {
		_, err := client.InspectImage(image)
		if err == dockerClient.ErrNoSuchImage {
			fmt.Println(image, "remote")
		} else {
			fmt.Println(image, "local")
		}
	}
}

func getLatestImage() (string, error) {
	images, err := getImages()
	if err != nil {
		return "", err
	}

	return images.Current, nil
}

func osUpgrade(c *cli.Context) {
	image := c.String("image")

	if image == "" {
		var err error
		image, err = getLatestImage()
		if err != nil {
			log.Fatal(err)
		}
		if image == "" {
			log.Fatal("Failed to find latest image")
		}
	}
	if c.Args().Present() {
		log.Fatalf("invalid arguments %v", c.Args())
	}
	if err := startUpgradeContainer(image, c.Bool("stage"), c.Bool("force"), !c.Bool("no-reboot"), c.Bool("kexec")); err != nil {
		log.Fatal(err)
	}
}

func osVersion(c *cli.Context) {
	fmt.Println(config.VERSION)
}

func yes(in *bufio.Reader, question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(line[0:1]) == "y"
}

func startUpgradeContainer(image string, stage, force, reboot, kexec bool) error {
	in := bufio.NewReader(os.Stdin)

	container, err := compose.CreateService(nil, "os-upgrade", &project.ServiceConfig{
		LogDriver:  "json-file",
		Privileged: true,
		Net:        "host",
		Image:      image,
		Labels: project.NewSliceorMap(map[string]string{
			config.SCOPE: config.SYSTEM,
		}),
		Command: project.NewCommand(
			"-t", "rancher-upgrade",
			"-r", config.VERSION,
		),
	})
	if err != nil {
		return err
	}

	if err := container.Pull(); err != nil {
		return err
	}

	if !stage {
		fmt.Printf("Upgrading to %s\n", image)

		if !force {
			if !yes(in, "Continue") {
				os.Exit(1)
			}
		}

		if err := container.Start(); err != nil {
			return err
		}

		if err := container.Log(); err != nil {
			return err
		}

		if err := container.Up(); err != nil {
			return err
		}

		if reboot && (force || yes(in, "Continue with reboot")) {
			if kexec {
				log.Info("Rebooting using kexec")

				version := strings.Split(image, ":")[1]
				vmlinuz := fmt.Sprintf("/boot/vmlinuz-%s-rancheros", version)
				initrd := fmt.Sprintf("--initrd=/boot/initrd-%s-rancheros", version)
				argv := []string{"kexec", "-l", vmlinuz, initrd, "-f"}

				return syscall.Exec("/sbin/kexec", argv, os.Environ())
			}

			log.Info("Rebooting")
			power.Reboot()
		}
	}

	return nil
}

func parseBody(body []byte) (*Images, error) {
	update := &Images{}
	err := yaml.Unmarshal(body, update)
	if err != nil {
		return nil, err
	}

	return update, nil
}

func getUpgradeUrl() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}

	return cfg.Rancher.Upgrade.Url, nil
}
