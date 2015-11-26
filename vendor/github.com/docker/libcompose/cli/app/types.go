package app

import (
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

// ProjectFactory is an interface that helps creating libcompose project.
type ProjectFactory interface {
	// Create creates a libcompose project from the command line options (codegangsta cli context).
	Create(c *cli.Context) (*project.Project, error)
}
