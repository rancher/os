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
	"gopkg.in/yaml.v2"

	dockerClient "github.com/fsouza/go-dockerclient"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/cmd/power"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
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
	startUpgradeContainer(image, c.Bool("stage"), c.Bool("force"))
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

func startUpgradeContainer(image string, stage, force bool) {
	in := bufio.NewReader(os.Stdin)

	container := docker.NewContainer(config.DOCKER_SYSTEM_HOST, &config.ContainerConfig{
		Cmd: "--name=os-upgrade " +
			"--rm " +
			"--privileged " +
			"--net=host " +
			image + " " +
			"-t rancher-upgrade " +
			"-r " + config.VERSION,
	}).Stage()

	if container.Err != nil {
		log.Fatal(container.Err)
	}

	fmt.Printf("Upgrading to %s\n", image)

	if !stage {
		if !force {
			if !yes(in, "Continue") {
				os.Exit(1)
			}
		}

		container.Start()
		if container.Err != nil {
			log.Fatal(container.Err)
		}

		client, err := docker.NewClient(config.DOCKER_SYSTEM_HOST)
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			client.Logs(dockerClient.LogsOptions{
				Container:    container.Container.ID,
				OutputStream: os.Stdout,
				ErrorStream:  os.Stderr,
				Follow:       true,
				Stdout:       true,
				Stderr:       true,
			})
		}()

		exit, err := client.WaitContainer(container.Container.ID)
		if err != nil {
			log.Fatal(err)
		}

		if container.Err != nil {
			log.Fatal(container.Err)
		}

		if exit == 0 {
			if force || yes(in, "Continue with reboot") {
				log.Info("Rebooting")
				power.Reboot()
			}
		} else {
			log.Error("Upgrade failed")
			os.Exit(exit)
		}
	}
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

	return cfg.Upgrade.Url, nil
}
