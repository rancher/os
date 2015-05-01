package docker

import (
	"strconv"
	"strings"

	"github.com/docker/docker/nat"
	"github.com/docker/docker/runconfig"
	"github.com/rancherio/rancher-compose/project"

	shlex "github.com/flynn/go-shlex"
)

func restartPolicy(restart string) runconfig.RestartPolicy {
	rs := strings.Split(restart, ":")
	result := runconfig.RestartPolicy{
		Name:               rs[0],
		MaximumRetryCount:  0,
	}
	if len(rs) == 2 {
		if i, err := strconv.Atoi(rs[1]); err == nil {
			result.MaximumRetryCount = i
		}
	}
	return result
}

func Convert(c *project.ServiceConfig) (*runconfig.Config, *runconfig.HostConfig, error) {

	cmd, _ := shlex.Split(c.Command)
	entrypoint, _ := shlex.Split(c.Entrypoint)
	ports, binding, err := nat.ParsePortSpecs(c.Ports)

	if err != nil {
		return nil, nil, err
	}

	config := &runconfig.Config{
		Entrypoint:   entrypoint,
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Memory:       c.MemLimit,
		CpuShares:    c.CpuShares,
		Env:          c.Environment,
		Cmd:          cmd,
		Image:        c.Image,
		Labels:       kvListToMap(c.Labels),
		ExposedPorts: ports,
	}
	host_config := &runconfig.HostConfig{
		VolumesFrom: c.VolumesFrom,
		CapAdd:      c.CapAdd,
		CapDrop:     c.CapDrop,
		Privileged:  c.Privileged,
		Binds:       c.Volumes,
		Dns:         c.Dns,
		LogConfig: runconfig.LogConfig{
			Type: c.LogDriver,
		},
		NetworkMode:    runconfig.NetworkMode(c.Net),
		ReadonlyRootfs: c.ReadOnly,
		PidMode:        runconfig.PidMode(c.Pid),
		IpcMode:        runconfig.IpcMode(c.Ipc),
		PortBindings:   binding,
		RestartPolicy:  restartPolicy(c.Restart),
	}

	return config, host_config, nil
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
