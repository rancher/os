package graph

import (
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/cliconfig"
	"github.com/docker/docker/pkg/streamformatter"
	"github.com/docker/docker/registry"
)

// ImagePushConfig stores push configuration.
type ImagePushConfig struct {
	// MetaHeaders store meta data about the image (DockerHeaders with prefix X-Meta- in the request).
	MetaHeaders map[string][]string
	// AuthConfig holds authentication information for authorizing with the registry.
	AuthConfig *cliconfig.AuthConfig
	// Tag is the specific variant of the image to be pushed, this tag used when image is pushed. If no tag is provided, all tags will be pushed.
	Tag string
	// OutStream is the output writer for showing the status of the push operation.
	OutStream io.Writer
}

// Pusher is an interface to define Push behavior.
type Pusher interface {
	// Push tries to push the image configured at the creation of Pusher.
	// Push returns an error if any, as well as a boolean that determines whether to retry Push on the next configured endpoint.
	//
	// TODO(tiborvass): have Push() take a reference to repository + tag, so that the pusher itself is repository-agnostic.
	Push() (fallback bool, err error)
}

// NewPusher returns a new instance of an implementation conforming to Pusher interface.
func (s *TagStore) NewPusher(endpoint registry.APIEndpoint, localRepo Repository, repoInfo *registry.RepositoryInfo, imagePushConfig *ImagePushConfig, sf *streamformatter.StreamFormatter) (Pusher, error) {
	switch endpoint.Version {
	case registry.APIVersion2:
		return &v2Pusher{
			TagStore:   s,
			endpoint:   endpoint,
			localRepo:  localRepo,
			repoInfo:   repoInfo,
			config:     imagePushConfig,
			sf:         sf,
			layersSeen: make(map[string]bool),
		}, nil
	case registry.APIVersion1:
		return &v1Pusher{
			TagStore:  s,
			endpoint:  endpoint,
			localRepo: localRepo,
			repoInfo:  repoInfo,
			config:    imagePushConfig,
			sf:        sf,
		}, nil
	}
	return nil, fmt.Errorf("unknown version %d for registry %s", endpoint.Version, endpoint.URL)
}

// FIXME: Allow to interrupt current push when new push of same image is done.

// Push a image to the repo.
func (s *TagStore) Push(localName string, imagePushConfig *ImagePushConfig) error {
	var sf = streamformatter.NewJSONStreamFormatter()

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := s.registryService.ResolveRepository(localName)
	if err != nil {
		return err
	}

	endpoints, err := s.registryService.LookupEndpoints(repoInfo.CanonicalName)
	if err != nil {
		return err
	}

	reposLen := 1
	if imagePushConfig.Tag == "" {
		reposLen = len(s.Repositories[repoInfo.LocalName])
	}

	imagePushConfig.OutStream.Write(sf.FormatStatus("", "The push refers to a repository [%s] (len: %d)", repoInfo.CanonicalName, reposLen))

	// If it fails, try to get the repository
	localRepo, exists := s.Repositories[repoInfo.LocalName]
	if !exists {
		return fmt.Errorf("Repository does not exist: %s", repoInfo.LocalName)
	}

	var lastErr error
	for _, endpoint := range endpoints {
		logrus.Debugf("Trying to push %s to %s %s", repoInfo.CanonicalName, endpoint.URL, endpoint.Version)

		pusher, err := s.NewPusher(endpoint, localRepo, repoInfo, imagePushConfig, sf)
		if err != nil {
			lastErr = err
			continue
		}
		if fallback, err := pusher.Push(); err != nil {
			if fallback {
				lastErr = err
				continue
			}
			logrus.Debugf("Not continuing with error: %v", err)
			return err

		}

		s.eventsService.Log("push", repoInfo.LocalName, "")
		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no endpoints found for %s", repoInfo.CanonicalName)
	}
	return lastErr
}
