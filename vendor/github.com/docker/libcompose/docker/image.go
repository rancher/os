package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

func removeImage(ctx context.Context, client client.APIClient, image string) error {
	_, err := client.ImageRemove(ctx, types.ImageRemoveOptions{
		ImageID: image,
	})
	return err
}

func pullImage(ctx context.Context, client client.APIClient, service *Service, image string) error {
	fmt.Fprintf(os.Stderr, "Pulling %s (%s)...\n", service.name, image)
	distributionRef, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}

	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	if err != nil {
		return err
	}

	authConfig := service.context.AuthLookup.Lookup(repoInfo)

	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	options := types.ImagePullOptions{
		ImageID:      distributionRef.String(),
		Tag:          "latest",
		RegistryAuth: encodedAuth,
	}
	if named, ok := distributionRef.(reference.Named); ok {
		options.ImageID = named.FullName()
	}
	if tagged, ok := distributionRef.(reference.NamedTagged); ok {
		options.Tag = tagged.Tag()
	}

	timeoutsRemaining := 3
	for i := 0; i < 100; i++ {
		responseBody, err := client.ImagePull(ctx, options, func() (string, error) {
			return encodedAuth, nil
		})
		if err != nil {
			logrus.Errorf("Failed to pull image %s: %v", image, err)
			return err
		}

		var writeBuff io.Writer = os.Stderr

		outFd, isTerminalOut := term.GetFdInfo(os.Stderr)

		err = jsonmessage.DisplayJSONMessagesStream(responseBody, writeBuff, outFd, isTerminalOut, nil)
		responseBody.Close()
		if err == nil {
			return nil
		} else if strings.Contains(err.Error(), "timed out") {
			timeoutsRemaining -= 1
			if timeoutsRemaining == 0 {
				return err
			}
			continue
		} else if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "unreachable") {
			time.Sleep(300 * time.Millisecond)
			continue
		} else {
			if jerr, ok := err.(*jsonmessage.JSONError); ok {
				// If no error code is set, default to 1
				if jerr.Code == 0 {
					jerr.Code = 1
				}
				fmt.Fprintf(os.Stderr, "%s", writeBuff)
				return fmt.Errorf("Status: %s, Code: %d", jerr.Message, jerr.Code)
			}
		}
	}

	return err
}

// encodeAuthToBase64 serializes the auth configuration as JSON base64 payload
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
