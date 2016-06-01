// +build !daemon

package docker

import "github.com/docker/docker/cli"

const daemonUsage = ""

var daemonCli cli.Handler

// notifySystem sends a message to the host when the server is ready to be used
func notifySystem() {
}
