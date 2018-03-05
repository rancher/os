package control

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/codegangsta/cli"

	dockerClient "github.com/docker/engine-api/client"
	"github.com/rancher/os/docker"
	"github.com/rancher/os/log"
)

const (
	userImagesPreloadDirectory = "/var/lib/rancher/preload/docker"
)

func preloadImagesAction(c *cli.Context) error {
	err := PreloadImages(docker.NewDefaultClient, userImagesPreloadDirectory)
	if err != nil {
		log.Errorf("Failed to preload user images: %v", err)
	}
	return err
}

func shouldLoad(file string) bool {
	if strings.HasSuffix(file, ".done") {
		return false
	}
	if _, err := os.Stat(fmt.Sprintf("%s.done", file)); err == nil {
		return false
	}
	return true
}

func PreloadImages(clientFactory func() (dockerClient.APIClient, error), imagesDir string) error {
	var client dockerClient.APIClient
	clientInitialized := false

	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		if err = os.MkdirAll(imagesDir, 0755); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(imagesDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := path.Join(imagesDir, file.Name())
		if !shouldLoad(filename) {
			log.Infof("Skipping to preload the file: %s", filename)
			continue
		}

		image, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer image.Close()
		var imageReader io.Reader
		imageReader = image
		match, err := regexp.MatchString(".t?gz$", file.Name())
		if err != nil {
			return err
		}
		if match {
			imageReader, err = gzip.NewReader(image)
			if err != nil {
				return err
			}
		}

		if !clientInitialized {
			client, err = clientFactory()
			if err != nil {
				return err
			}
			clientInitialized = true
		}

		log.Infof("Loading image %s", filename)
		if _, err = client.ImageLoad(context.Background(), imageReader, false); err != nil {
			return err
		}
		log.Infof("Finished to load image %s", filename)

		log.Infof("Creating done stamp file for image %s", filename)
		doneStamp, err := os.Create(fmt.Sprintf("%s.done", filename))
		if err != nil {
			return err
		}
		defer doneStamp.Close()
		log.Infof("Finished to created the done stamp file for image %s", filename)
	}

	return nil
}
