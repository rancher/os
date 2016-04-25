package control

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

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
				cli.StringFlag{
					Name:  "append",
					Usage: "kernel args to append by kexec",
				},
				cli.BoolFlag{
					Name:  "upgrade-console",
					Usage: "upgrade console even if persistent",
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
	if err := startUpgradeContainer(image, c.Bool("stage"), c.Bool("force"), !c.Bool("no-reboot"), c.Bool("kexec"), c.Bool("upgrade-console"), c.String("append")); err != nil {
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

func startUpgradeContainer(image string, stage, force, reboot, kexec bool, upgradeConsole bool, kernelArgs string) error {
	in := bufio.NewReader(os.Stdin)

	command := []string{
		"-t", "rancher-upgrade",
		"-r", config.VERSION,
	}

	if kexec {
		command = append(command, "-k")

		kernelArgs = strings.TrimSpace(kernelArgs)
		if kernelArgs != "" {
			command = append(command, "-a", kernelArgs)
		}
	}

	if upgradeConsole {
		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}

		cfg.Rancher.ForceConsoleRebuild = true
		if err := cfg.Save(); err != nil {
			log.Fatal(err)
		}
	}

	container, err := compose.CreateService(nil, "os-upgrade", &project.ServiceConfig{
		LogDriver:  "json-file",
		Privileged: true,
		Net:        "host",
		Pid:        "host",
		Image:      image,
		Labels: project.NewSliceorMap(map[string]string{
			config.SCOPE: config.SYSTEM,
		}),
		Command: project.NewCommand(command...),
	})
	if err != nil {
		return err
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}

	// Only pull image if not found locally
	if _, err := client.InspectImage(image); err != nil {
		if err := container.Pull(); err != nil {
			return err
		}
	}

	if !stage {
		imageSplit := strings.Split(image, ":")
		if len(imageSplit) > 1 && imageSplit[1] == config.VERSION {
			if !force && !yes(in, fmt.Sprintf("Already at version %s. Continue anyways", imageSplit[1])) {
				os.Exit(1)
			}
		} else {
			fmt.Printf("Upgrading to %s\n", image)

			if !force && !yes(in, "Continue") {
				os.Exit(1)
			}
		}

		// If there is already an upgrade container, delete it
		// Up() should to this, but currently does not due to a bug
		if err := container.Delete(); err != nil {
			return err
		}

		if err := container.Up(); err != nil {
			return err
		}

		if err := container.Log(); err != nil {
			return err
		}

		if err := container.Delete(); err != nil {
			return err
		}

		if reboot && (force || yes(in, "Continue with reboot")) {
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
