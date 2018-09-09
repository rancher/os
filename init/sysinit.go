package init

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/types"
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

func getImagesArchive(bootstrap bool) string {
	var archive string
	if bootstrap {
		archive = path.Join(config.ImagesPath, config.InitImages)
	} else {
		archive = path.Join(config.ImagesPath, config.SystemImages)
	}

	return archive
}

func loadBootstrapImages(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	return loadImages(cfg, true)
}

func loadSystemImages(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	return loadImages(cfg, false)
}

func loadImages(cfg *config.CloudConfig, bootstrap bool) (*config.CloudConfig, error) {
	archive := getImagesArchive(bootstrap)

	client, err := docker.NewSystemClient()
	if err != nil {
		return cfg, err
	}

	if !hasImage(filepath.Base(archive)) {
		if _, err := os.Stat(archive); os.IsNotExist(err) {
			log.Fatalf("FATAL: Could not load images from %s (file not found)", archive)
		}

		// client.ImageLoad is an asynchronous operation
		// To ensure the order of execution, use cmd instead of it
		log.Infof("Loading images from %s", archive)
		cmd := exec.Command("/usr/bin/system-docker", "load", "-q", "-i", archive)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Fatalf("FATAL: Error loading images from %s (%v)\n%s ", archive, err, out)
		}

		log.Infof("Done loading images from %s", archive)
	}

	dockerImages, _ := client.ImageList(context.Background(), types.ImageListOptions{})
	for _, dimg := range dockerImages {
		log.Infof("Got image repo tags: %s", dimg.RepoTags)
	}

	return cfg, nil
}

func SysInit() error {
	cfg := config.LoadConfig()

	if err := control.PreloadImages(docker.NewSystemClient, systemImagesPreloadDirectory); err != nil {
		log.Errorf("Failed to preload System Docker images: %v", err)
	}

	_, err := config.ChainCfgFuncs(cfg,
		[]config.CfgFuncData{
			config.CfgFuncData{"loadSystemImages", loadSystemImages},
			config.CfgFuncData{"start project", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
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
			}},
			config.CfgFuncData{"sync", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				syscall.Sync()
				return cfg, nil
			}},
			config.CfgFuncData{"banner", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				log.Infof("RancherOS %s started", config.Version)
				return cfg, nil
			}}})
	return err
}
