package app

import (
	"github.com/codegangsta/cli"
)

// DockerClientFlags defines the flags that are specific to the docker client,
// like configdir or tls related flags.
func DockerClientFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "tls",
			Usage: "Use TLS; implied by --tlsverify",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "Use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:  "tlscacert",
			Usage: "Trust certs signed only by this CA",
		},
		cli.StringFlag{
			Name:  "tlscert",
			Usage: "Path to TLS certificate file",
		},
		cli.StringFlag{
			Name:  "tlskey",
			Usage: "Path to TLS key file",
		},
		cli.StringFlag{
			Name:  "configdir",
			Usage: "Path to docker config dir, default ${HOME}/.docker",
		},
	}
}
