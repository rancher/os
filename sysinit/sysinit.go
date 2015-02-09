package sysinit

import (
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
	initPkg "github.com/rancherio/os/init"
)

func SysInit() {
	if err := sysInit(); err != nil {
		log.Fatal(err)
	}
}

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
	stamp := path.Join(initPkg.STATE, name)
	if _, err := os.Stat(stamp); os.IsNotExist(err) {
		return false
	}
	return true
}

func findImages(cfg *config.Config) ([]string, error) {
	log.Debugf("Looking for images at %s", cfg.ImagesPath)

	result := []string{}

	dir, err := os.Open(cfg.ImagesPath)
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
		log.Debugf("Checking %s", fileName)
		if ok, _ := path.Match(cfg.ImagesPattern, fileName); ok {
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

	client, err := docker.NewClient(cfg)
	if err != nil {
		return err
	}

	for _, image := range images {
		if hasImage(image) {
			continue
		}

		inputFileName := path.Join(cfg.ImagesPath, image)
		input, err := os.Open(inputFileName)
		if err != nil {
			return err
		}

		defer input.Close()

		log.Debugf("Loading images from %s", inputFileName)
		err = client.LoadImage(dockerClient.LoadImageOptions{
			InputStream: input,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func runContainers(cfg *config.Config) error {
	containers := cfg.SystemContainers
	if cfg.Rescue {
		log.Debug("Running rescue container")
		containers = []config.ContainerConfig{cfg.RescueContainer}
	}

	for _, container := range containers {
		args := append([]string{"run"}, container.Options...)
		args = append(args, container.Image)
		args = append(args, container.Args...)

		cmd := exec.Command(cfg.DockerBin, args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		//log.Infof("Is a tty : %v", term.IsTerminal(0))
		//log.Infof("Is a tty : %v", term.IsTerminal(1))
		//log.Infof("Is a tty : %v", term.IsTerminal(2))
		log.Debugf("Running %s", strings.Join(args, " "))
		err := cmd.Run()
		if err != nil {
			log.Errorf("Failed to run %v: %v", args, err)
		}
	}

	return nil
}

func sysInit() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	initFuncs := []config.InitFunc{
		loadImages,
		runContainers,
	}

	return config.RunInitFuncs(cfg, initFuncs)
}
