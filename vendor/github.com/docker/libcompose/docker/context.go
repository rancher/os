package docker

import (
	"github.com/docker/docker/cliconfig"
	"github.com/docker/libcompose/project"
)

type Context struct {
	project.Context
	Builder       Builder
	ClientFactory ClientFactory
	ConfigDir     string
	ConfigFile    *cliconfig.ConfigFile
}

func (c *Context) open() error {
	return c.LookupConfig()
}

func (c *Context) LookupConfig() error {
	if c.ConfigFile != nil {
		return nil
	}

	config, err := cliconfig.Load(c.ConfigDir)
	if err != nil {
		return err
	}

	c.ConfigFile = config

	return nil
}
