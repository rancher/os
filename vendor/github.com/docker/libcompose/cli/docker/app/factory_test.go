package app

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

func TestProjectFactoryProjectNameIsNormalized(t *testing.T) {
	projects := []struct {
		name     string
		expected string
	}{
		{
			name:     "example",
			expected: "example",
		},
		{
			name:     "example-test",
			expected: "exampletest",
		},
		{
			name:     "aW3Ird_Project_with_$$",
			expected: "aw3irdprojectwith",
		},
	}

	tmpDir, err := ioutil.TempDir("", "project-factory-test")
	if err != nil {
		t.Fatal(err)
	}
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	ioutil.WriteFile(composeFile, []byte(`hello:
    image: busybox`), 0700)

	for _, projectCase := range projects {
		globalSet := flag.NewFlagSet("test", 0)
		// Set the project-name flag
		globalSet.String("project-name", projectCase.name, "doc")
		// Set the compose file flag
		globalSet.Var(&cli.StringSlice{composeFile}, "file", "doc")
		c := cli.NewContext(nil, globalSet, nil)
		factory := &ProjectFactory{}
		p, err := factory.Create(c)
		if err != nil {
			t.Fatal(err)
		}

		if p.(*project.Project).Name != projectCase.expected {
			t.Fatalf("expected %s, got %s", projectCase.expected, p.(*project.Project).Name)
		}
	}
}
