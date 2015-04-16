package init

import (
	"os"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
)

func importImage(client *dockerClient.Client, name, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	log.Debugf("Importing image for %s", fileName)
	repo, tag := dockerClient.ParseRepositoryTag(name)
	return client.ImportImage(dockerClient.ImportImageOptions{
		Source:      "-",
		Repository:  repo,
		Tag:         tag,
		InputStream: file,
	})
}

func hasImage(name string) bool {
	stamp := path.Join(STATE, name)
	if _, err := os.Stat(stamp); os.IsNotExist(err) {
		return false
	}
	return true
}

func findImages(cfg *config.Config) ([]string, error) {
	log.Debugf("Looking for images at %s", config.IMAGES_PATH)

	result := []string{}

	dir, err := os.Open(config.IMAGES_PATH)
	if os.IsNotExist(err) {
		log.Debugf("Not loading images, %s does not exist")
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	for _, fileName := range files {
		if ok, _ := path.Match(config.IMAGES_PATTERN, fileName); ok {
			log.Debugf("Found %s", fileName)
			result = append(result, fileName)
		}
	}

	return result, nil
}

func loadImages(cfg *config.Config) error {
	images, err := findImages(cfg)
	if err != nil || len(images) == 0 {
		return err
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}

	for _, image := range images {
		if hasImage(image) {
			continue
		}

		inputFileName := path.Join(config.IMAGES_PATH, image)
		input, err := os.Open(inputFileName)
		if err != nil {
			return err
		}

		defer input.Close()

		log.Infof("Loading images from %s", inputFileName)
		err = client.LoadImage(dockerClient.LoadImageOptions{
			InputStream: input,
		})
		log.Infof("Done loading images from %s", inputFileName)

		if err != nil {
			return err
		}
	}

	return nil
}

func runContainers(cfg *config.Config) error {
	return docker.RunServices("system-init", cfg, cfg.SystemContainers)
}

func tailConsole(cfg *config.Config) error {
	if !cfg.Console.Tail {
		return nil
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}

	console, ok := cfg.SystemContainers[config.CONSOLE_CONTAINER]
	if !ok {
		log.Error("Console not found")
		return nil
	}

	c := docker.NewContainerFromService(config.DOCKER_SYSTEM_HOST, config.CONSOLE_CONTAINER, console)
	if c.Err != nil {
		return c.Err
	}

	log.Infof("Tailing console : %s", c.Name)
	return client.Logs(dockerClient.LogsOptions{
		Container:    c.Name,
		Stdout:       true,
		Stderr:       true,
		Follow:       true,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
	})
}

func SysInit() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	initFuncs := []config.InitFunc{
		loadImages,
		runContainers,
		func(cfg *config.Config) error {
			syscall.Sync()
			return nil
		},
		func(cfg *config.Config) error {
			log.Infof("RancherOS %s started", config.VERSION)
			return nil
		},
		tailConsole,
	}

	return config.RunInitFuncs(cfg, initFuncs)
}
