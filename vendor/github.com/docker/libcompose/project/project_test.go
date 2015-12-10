package project

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type TestServiceFactory struct {
	Counts map[string]int
}

type TestService struct {
	factory *TestServiceFactory
	name    string
	config  *ServiceConfig
	EmptyService
	Count int
}

func (t *TestService) Config() *ServiceConfig {
	return t.config
}

func (t *TestService) Name() string {
	return t.name
}

func (t *TestService) Create() error {
	key := t.name + ".create"
	t.factory.Counts[key] = t.factory.Counts[key] + 1
	return nil
}

func (t *TestService) DependentServices() []ServiceRelationship {
	return nil
}

func (t *TestServiceFactory) Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error) {
	return &TestService{
		factory: t,
		config:  serviceConfig,
		name:    name,
	}, nil
}

func TestTwoCall(t *testing.T) {
	factory := &TestServiceFactory{
		Counts: map[string]int{},
	}

	p := NewProject(&Context{
		ServiceFactory: factory,
	})
	p.Configs = map[string]*ServiceConfig{
		"foo": {},
	}

	if err := p.Create("foo"); err != nil {
		t.Fatal(err)
	}

	if err := p.Create("foo"); err != nil {
		t.Fatal(err)
	}

	if factory.Counts["foo.create"] != 2 {
		t.Fatal("Failed to create twice")
	}
}

func TestEventEquality(t *testing.T) {
	if fmt.Sprintf("%s", EventServiceStart) != "Started" ||
		fmt.Sprintf("%v", EventServiceStart) != "Started" {
		t.Fatalf("EventServiceStart String() doesn't work: %s %v", EventServiceStart, EventServiceStart)
	}

	if fmt.Sprintf("%s", EventServiceStart) != fmt.Sprintf("%s", EventServiceUp) {
		t.Fatal("Event messages do not match")
	}

	if EventServiceStart == EventServiceUp {
		t.Fatal("Events match")
	}
}

func TestParseWithBadContent(t *testing.T) {
	p := NewProject(&Context{
		ComposeBytes: []byte("garbage"),
	})

	err := p.Parse()
	if err == nil {
		t.Fatal("Should have failed parse")
	}

	if !strings.HasPrefix(err.Error(), "Unknown resolution for 'garbage'") {
		t.Fatalf("Should have failed parse: %#v", err)
	}
}

func TestParseWithGoodContent(t *testing.T) {
	p := NewProject(&Context{
		ComposeBytes: []byte("not-garbage:\n  image: foo"),
	})

	err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
}

type TestEnvironmentLookup struct {
}

func (t *TestEnvironmentLookup) Lookup(key, serviceName string, config *ServiceConfig) []string {
	return []string{fmt.Sprintf("%s=X", key)}
}

func TestEnvironmentResolve(t *testing.T) {
	factory := &TestServiceFactory{
		Counts: map[string]int{},
	}

	p := NewProject(&Context{
		ServiceFactory:    factory,
		EnvironmentLookup: &TestEnvironmentLookup{},
	})
	p.Configs = map[string]*ServiceConfig{
		"foo": {
			Environment: NewMaporEqualSlice([]string{
				"A",
				"A=",
				"A=B",
			}),
		},
	}

	service, err := p.CreateService("foo")
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(service.Config().Environment.Slice(), []string{"A=X", "A=X", "A=B"}) {
		t.Fatal("Invalid environment", service.Config().Environment.Slice())
	}
}
