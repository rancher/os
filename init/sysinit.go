package init

import (
	"os"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
)

func hasImage(name string) bool {
	stamp := path.Join(STATE, name)
	if _, err := os.Stat(stamp); os.IsNotExist(err) {
		return false
	}
	return true
}

func findImages(cfg *config.CloudConfig) ([]string, error) {
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

func loadImages(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	images, err := findImages(cfg)
	if err != nil || len(images) == 0 {
		return cfg, err
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return cfg, err
	}

	for _, image := range images {
		if hasImage(image) {
			continue
		}

		inputFileName := path.Join(config.IMAGES_PATH, image)
		input, err := os.Open(inputFileName)
		if err != nil {
			return cfg, err
		}

		defer input.Close()

		log.Infof("Loading images from %s", inputFileName)
		err = client.LoadImage(dockerClient.LoadImageOptions{
			InputStream: input,
		})
		if err != nil {
			return cfg, err
		}

		log.Infof("Done loading images from %s", inputFileName)
	}

	return cfg, nil
}

func SysInit() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	_, err = config.ChainCfgFuncs(cfg,
		loadImages,
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			p, err := compose.GetProject(cfg, false)
			if err != nil {
				return cfg, err
			}
			return cfg, p.Up()
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			syscall.Sync()
			return cfg, nil
		},
		func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
			log.Infof("RancherOS %s started", config.VERSION)
			return cfg, nil
		})
	return err
}
