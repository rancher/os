package config

import "os"

func (d *DockerConfig) FullArgs() []string {
	args := append(d.Args, d.ExtraArgs...)

	if d.TLS {
		args = append(args, d.TLSArgs...)
	}

	return args
}

func (d *DockerConfig) AppendEnv() []string {
	return append(os.Environ(), d.Environment...)
}
