package sysinit

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"syscall"

	"github.com/docker/engine-api/types"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/compose"
	"github.com/rancher/os/pkg/docker"
	"github.com/rancher/os/pkg/log"
	"golang.org/x/net/context"
)

const (
	systemImagesPreloadDirectory = "/var/lib/rancher/preload/system-docker"
)

func hasImage(name string) bool {
	stamp := path.Join(config.StateDir, name)
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

func LoadBootstrapImages(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	return loadImages(cfg, true)
}

func LoadSystemImages(cfg *config.CloudConfig) (*config.CloudConfig, error) {
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
		config.CfgFuncs{
			{"loadSystemImages", LoadSystemImages},
			{"start project", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
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
			{"sync", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				syscall.Sync()
				return cfg, nil
			}},
			{"banner", func(cfg *config.CloudConfig) (*config.CloudConfig, error) {
				log.Infof("RancherOS %s started", config.Version)
				return cfg, nil
			}}})
	return err
}

func RunSysInit(c *config.CloudConfig) (*config.CloudConfig, error) {
	args := append([]string{config.SysInitBin}, os.Args[1:]...)

	cmd := &exec.Cmd{
		Path: config.RosBin,
		Args: args,
	}

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return c, err
	}

	return c, os.Stdin.Close()
}
