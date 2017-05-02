package init

import (
	"os"
	"path"
	"syscall"

	"golang.org/x/net/context"

	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"
)

const (
	systemImagesPreloadDirectory = "/var/lib/rancher/preload/system-docker"
)

func hasImage(name string) bool {
	stamp := path.Join(state, name)
	if _, err := os.Stat(stamp); os.IsNotExist(err) {
		return false
	}
	return true
}

func findImages(cfg *config.CloudConfig) ([]string, error) {
	log.Debugf("Looking for images at %s", config.ImagesPath)

	result := []string{}

	dir, err := os.Open(config.ImagesPath)
	if os.IsNotExist(err) {
		log.Debugf("Not loading images, %s does not exist", config.ImagesPath)
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
		if ok, _ := path.Match(config.ImagesPattern, fileName); ok {
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

		inputFileName := path.Join(config.ImagesPath, image)
		input, err := os.Open(inputFileName)
		if err != nil {
			return cfg, err
		}

		defer input.Close()

		log.Infof("Loading images from %s", inputFileName)
		if _, err = client.ImageLoad(context.Background(), input, true); err != nil {
			log.Fatalf("FATAL: failed loading images from %s: %s", inputFileName, err)
		}

		log.Infof("Done loading images from %s", inputFileName)
	}

	return cfg, nil
}

func SysInit() error {
	cfg := config.LoadConfig()

	if err := control.PreloadImages(docker.NewSystemClient, systemImagesPreloadDirectory); err != nil {
		log.Errorf("Failed to preload System Docker images: %v", err)
	}

	_, err := config.ChainCfgFuncs(cfg,
		map[string]config.CfgFunc{
			"loadImages": loadImages,
			"start project": func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				p, err := compose.GetProject(cfg, false, true)
				if err != nil {
					return cfg, err
				}
				return cfg, p.Up(context.Background(), options.Up{
					Create: options.Create{
						NoRecreate: true,
					},
					Log: cfg.Rancher.Log,
				})
			},
			"sync": func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				syscall.Sync()
				return cfg, nil
			},
			"banner": func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				log.Infof("RancherOS %s started", config.Version)
				return cfg, nil
			}})
	return err
}
