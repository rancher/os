package docker

import (
	"github.com/docker/libcompose/project"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSpecifiesHostPort(t *testing.T) {
	servicesWithHostPort := []Service{
		{serviceConfig: &project.ServiceConfig{Ports: []string{"8000:8000"}}},
		{serviceConfig: &project.ServiceConfig{Ports: []string{"127.0.0.1:8000:8000"}}},
	}

	for _, service := range servicesWithHostPort {
		assert.True(t, service.specificiesHostPort())
	}

	servicesWithoutHostPort := []Service{
		{serviceConfig: &project.ServiceConfig{Ports: []string{"8000"}}},
		{serviceConfig: &project.ServiceConfig{Ports: []string{"127.0.0.1::8000"}}},
	}

	for _, service := range servicesWithoutHostPort {
		assert.False(t, service.specificiesHostPort())
	}
}
