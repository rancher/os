package project

import "testing"

type testFactory struct {
}

func (t *testFactory) Create(project *Project, name string, config *ServiceConfig) (Service, error) {
	return struct{}{}, nil
}

func TestNewProject(t *testing.T) {
	p, err := NewProject("foo", "test_files/docker-compose.yml", &testFactory{})
	if err != nil {
		t.Fatal(err)
	}

	if p.Name != "foo" {
		t.Fatal("Wrong name expected foo, got", p.Name)
	}
}
