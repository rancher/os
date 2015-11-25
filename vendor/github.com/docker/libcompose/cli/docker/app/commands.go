package app

import (
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/docker"
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

// Populate updates the specified docker context based on command line arguments and subcommands.
func Populate(context *docker.Context, c *cli.Context) {
	context.ConfigDir = c.String("configdir")

	opts := docker.ClientOpts{}
	opts.TLS = c.GlobalBool("tls")
	opts.TLSVerify = c.GlobalBool("tlsverify")
	opts.TLSOptions.CAFile = c.GlobalString("tlscacert")
	opts.TLSOptions.CertFile = c.GlobalString("tlscert")
	opts.TLSOptions.KeyFile = c.GlobalString("tlskey")

	clientFactory, err := docker.NewDefaultClientFactory(opts)
	if err != nil {
		logrus.Fatalf("Failed to construct Docker client: %v", err)
	}

	context.ClientFactory = clientFactory
}
