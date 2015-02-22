package control

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/rancherio/os/cmd/power"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

var osChannels map[string]string

const (
	osVersionsFile = "/var/lib/rancher/versions"
)

func osSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "upgrade",
			Usage:  "upgrade to latest version",
			Action: osUpgrade,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "stage, s",
					Usage: "Only stage the new upgrade, don't apply it",
				},
				cli.StringFlag{
					Name: "image, i",
					Usage: "upgrade to a certain image",
				},
				cli.StringFlag{
					Name: "channel, c",
					Usage: "upgrade to the latest in a specific channel",
				},
			},
		},
		{
			Name: "list",
			Usage: "list the current available versions",
			Action: osMetaDataGet,
		},
		{
			Name: "rollback",
			Usage: "rollback to the previous version",
			Action: osRollback,
		},
	}
}

func osRollback(c *cli.Context) {
	file, err := os.Open(osVersionsFile)

	if err != nil {
		log.Fatal(err)
	}

	fileReader := bufio.NewScanner(file)
	line := " "
	for ; line[len(line)-1:] != "*"; {
		if !fileReader.Scan() {
			log.Error("Current version not indicated in "+ osVersionsFile)
		}
		line = fileReader.Text()
	}
	if !fileReader.Scan() {
		log.Error("already at earliest version, please choose a version specifically using upgrade --image")
	}
	line = fileReader.Text()
	//TODO: process string if required

	startUpgradeContainer(line, false)
}
	
func osMetaDataGet(c *cli.Context) {
	osChannel, ok := getChannelUrl("meta"); if !ok {
		log.Fatal("unrecognized channel meta")
	}
	resp, err := http.Get(osChannel)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(parseBody(body, osChannel))
}

func osUpgrade(c *cli.Context) {
	channel := c.String("channel")

	image := c.String("image")

	if image == "" {
		var err error
		image, err = getLatestImage(channel)
		if err != nil {
			log.Fatal(err)
		}
	}
	startUpgradeContainer(image, c.Bool("stage"))
}

func startUpgradeContainer(image string, stage bool) {
	container := docker.NewContainer(config.DOCKER_SYSTEM_HOST, &config.ContainerConfig{
		Cmd:  "--name=upgrade " +
			"--privileged " +
			"--net=host " +
			"--ipc=host " +
			"--pid=host " +
			"-v=/var:/var " +
			"--volumes-from=system-volumes " +
			image,
	}).Stage()

	if container.Err != nil {
		log.Fatal(container.Err)
	}

	if !stage {
		container.StartAndWait()
		if container.Err != nil {
			log.Fatal(container.Err)
		}
		power.Reboot()
	}
}

func getLatestImage(channel string) (string, error) {
	data, err := getConfigData()

	if err != nil {
		return "", err	
	}

	var pivot string

	if pivot == "" {
		val := getOrSetVal("os_upgrade_channel", data, nil)

		if val == nil {
			return "", errors.New("os_upgrade_channel is not set")
		}

		switch currentChannel := val.(type) {
			case string:
				pivot = currentChannel
			default:
				return "", errors.New("invalid format of rancherctl config get os_upgrade_channel")
		}
	} else {
		pivot = channel
	}
	osChannel, ok := getChannelUrl(pivot); if !ok {
		return "", errors.New("unrecognized channel " + pivot)
	}
	resp, err := http.Get(osChannel)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return parseBody(body, osChannel), nil
}

func parseBody(body []byte, channel string) string {
	// just going to assume that the response is the image name
	// can change it later based on server response design
	return string(body)
}

func getChannelUrl(channel string) (string, bool) {
	if osChannels == nil {
		osChannels = map[string]string {
				"stable" : "",
				"alpha" : "",
				"beta" : "",
				"meta" : "",
		}
	}
	channel, ok := osChannels[channel]; 
	return channel, ok
}

