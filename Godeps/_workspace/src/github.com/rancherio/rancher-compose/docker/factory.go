package docker

import (
	"strings"

	"github.com/docker/docker/runconfig"
	"github.com/rancherio/rancher-compose/project"

	shlex "github.com/flynn/go-shlex"
)

func Convert(c *project.ServiceConfig) (*runconfig.Config, *runconfig.HostConfig, error) {
	volumes := map[string]struct{}{}
	for _, v := range c.Volumes {
		volumes[strings.Split(v, ":")[0]] = struct{}{}
	}

	cmd, _ := shlex.Split(c.Command)

	return &runconfig.Config{
			Hostname:   c.Hostname,
			Domainname: c.DomainName,
			User:       c.User,
			Memory:     c.MemLimit,
			CpuShares:  c.CpuShares,
			Env:        c.Environment,
			Cmd:        cmd,
			Image:      c.Image,
			Labels:     kvListToMap(c.Labels),
		},
		&runconfig.HostConfig{
			VolumesFrom:    c.VolumesFrom,
			CapAdd:         c.CapAdd,
			CapDrop:        c.CapDrop,
			Privileged:     c.Privileged,
			Binds:          c.Volumes,
			Dns:            c.Dns,
			NetworkMode:    runconfig.NetworkMode(c.Net),
			ReadonlyRootfs: c.ReadOnly,
			PidMode:        runconfig.PidMode(c.Pid),
			IpcMode:        runconfig.IpcMode(c.Ipc),
		},
		nil
}

func kvListToMap(list []string) map[string]string {
	result := make(map[string]string)
	for _, item := range list {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) < 2 {
			result[parts[0]] = ""
		} else {
			result[parts[0]] = parts[1]
		}
	}

	return result
}
