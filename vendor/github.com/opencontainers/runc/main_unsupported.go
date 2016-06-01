// +build !linux

package runc

import "github.com/codegangsta/cli"

var (
	checkpointCommand cli.Command
	eventsCommand     cli.Command
	restoreCommand    cli.Command
	specCommand       cli.Command
	killCommand       cli.Command
)
