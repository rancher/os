package docker

import (
	"strings"

	"github.com/docker/docker/nat"
	"github.com/docker/docker/runconfig"
	"github.com/rancherio/rancher-compose/librcompose/project"

	shlex "github.com/flynn/go-shlex"
)

func Filter(vs []string, f func(string) bool) []string {
	r := make([]string, 0, len(vs))
	for _, v := range vs {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func isBind(s string) bool {
	return strings.ContainsRune(s, ':')
}

func isVolume(s string) bool {
	return !isBind(s)
}

func Convert(c *project.ServiceConfig) (*runconfig.Config, *runconfig.HostConfig, error) {
	vs := Filter(c.Volumes, isVolume)

	volumes := make(map[string]struct {}, len(vs))
	for _, v := range vs {
		volumes[v] = struct {}{}
	}

	cmd, _ := shlex.Split(c.Command)
	entrypoint, _ := shlex.Split(c.Entrypoint)
	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}
	restart, err := runconfig.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, nil, err
	}
	dns := c.Dns.Slice()
	dnssearch := c.DnsSearch.Slice()
	labels := c.Labels.MapParts()

	if len(c.Expose) > 0 {
		exposedPorts, _, err := nat.ParsePortSpecs(c.Expose)
		ports = exposedPorts
		if err != nil {
			return nil, nil, err
		}
	}

	config := &runconfig.Config{
		Entrypoint:   runconfig.NewEntrypoint(entrypoint...),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          c.Environment.Slice(),
		Cmd:          runconfig.NewCommand(cmd...),
		Image:        c.Image,
		Labels:       labels,
		ExposedPorts: ports,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		Volumes:      volumes,
	}
	host_config := &runconfig.HostConfig{
		VolumesFrom: c.VolumesFrom,
		CapAdd:      c.CapAdd,
		CapDrop:     c.CapDrop,
		CpuShares:   c.CpuShares,
		Privileged:  c.Privileged,
		Binds:       Filter(c.Volumes, isBind),
		Dns:         dns,
		DnsSearch:   dnssearch,
		LogConfig:   runconfig.LogConfig{
			Type: c.LogDriver,
		},
		Memory:         c.MemLimit,
		NetworkMode:    runconfig.NetworkMode(c.Net),
		ReadonlyRootfs: c.ReadOnly,
		PidMode:        runconfig.PidMode(c.Pid),
		IpcMode:        runconfig.IpcMode(c.Ipc),
		PortBindings:   binding,
		RestartPolicy:  restart,
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
