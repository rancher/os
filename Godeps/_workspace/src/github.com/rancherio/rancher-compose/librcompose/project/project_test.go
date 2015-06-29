package project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testFactory struct {
}

func (*testFactory) Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error) {
	return &testService{}, nil
}

type testService struct {
	EmptyService
}

func (*testService) Name() string           { return "" }
func (*testService) Up() error              { return nil }
func (*testService) Config() *ServiceConfig { return &ServiceConfig{} }

func TestNewProject(t *testing.T) {
	p := NewProject("foo", &testFactory{})
	assert.Equal(t, "foo", p.Name)
}
