package app

import (
	"github.com/burmilla/os/pkg/libcompose/project"
	"github.com/codegangsta/cli"
	"github.com/sirupsen/logrus"
)

// ProjectAction is an adapter to allow the use of ordinary functions as libcompose actions.
// Any function that has the appropriate signature can be register as an action on a codegansta/cli command.
//
//	cli.Command{
//			Name:   "ps",
//			Usage:  "List containers",
//			Action: app.WithProject(factory, app.ProjectPs),
//		}
type ProjectAction func(project project.APIProject, c *cli.Context) error

// WithProject is a helper function to create a cli.Command action with a ProjectFactory.
func WithProject(factory ProjectFactory, action ProjectAction) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		p, err := factory.Create(context)
		if err != nil {
			logrus.Fatalf("Failed to read project: %v", err)
		}
		return action(p, context)
	}
}
