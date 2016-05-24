package project

import (
	"github.com/docker/engine-api/client"
	composeclient "github.com/docker/libcompose/docker/client"
)

// ClientFactory is a factory to create docker clients.
type ClientFactory interface {
	// Create constructs a Docker client for the given service. The passed in
	// config may be nil in which case a generic client for the project should
	// be returned.
	Create(service Service) client.APIClient
}

type defaultClientFactory struct {
	client client.APIClient
}

// NewDefaultClientFactory creates and returns the default client factory that uses
// github.com/docker/engine-api client.
func NewDefaultClientFactory(opts composeclient.Options) (ClientFactory, error) {
	client, err := composeclient.Create(opts)
	if err != nil {
		return nil, err
	}

	return &defaultClientFactory{
		client: client,
	}, nil
}

func (s *defaultClientFactory) Create(service Service) client.APIClient {
	return s.client
}
