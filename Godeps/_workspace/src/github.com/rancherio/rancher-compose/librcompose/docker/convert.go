package docker

import (
	"strings"

	"github.com/docker/docker/nat"
	"github.com/docker/docker/runconfig"
	"github.com/rancherio/rancher-compose/librcompose/project"
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

	volumes := make(map[string]struct{}, len(vs))
	for _, v := range vs {
		volumes[v] = struct{}{}
	}

	ports, binding, err := nat.ParsePortSpecs(c.Ports)
	if err != nil {
		return nil, nil, err
	}
	restart, err := runconfig.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, nil, err
	}

	if exposedPorts, _, err := nat.ParsePortSpecs(c.Expose); err != nil {
		return nil, nil, err
	} else {
		for k, v := range exposedPorts {
			ports[k] = v
		}
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return nil, nil, err
	}

	config := &runconfig.Config{
		Entrypoint:   runconfig.NewEntrypoint(c.Entrypoint.Slice()...),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          c.Environment.Slice(),
		Cmd:          runconfig.NewCommand(c.Command.Slice()...),
		Image:        c.Image,
		Labels:       c.Labels.MapParts(),
		ExposedPorts: ports,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		VolumeDriver: c.VolumeDriver,
		Volumes:      volumes,
	}
	host_config := &runconfig.HostConfig{
		VolumesFrom: c.VolumesFrom,
		CapAdd:      c.CapAdd,
		CapDrop:     c.CapDrop,
		CpuShares:   c.CpuShares,
		CpusetCpus:  c.CpuSet,
		ExtraHosts:  c.ExtraHosts,
		Privileged:  c.Privileged,
		Binds:       Filter(c.Volumes, isBind),
		Devices:     deviceMappings,
		Dns:         c.Dns.Slice(),
		DnsSearch:   c.DnsSearch.Slice(),
		LogConfig: runconfig.LogConfig{
			Type:   c.LogDriver,
			Config: c.LogOpt,
		},
		Memory:         c.MemLimit,
		MemorySwap:     c.MemSwapLimit,
		NetworkMode:    runconfig.NetworkMode(c.Net),
		ReadonlyRootfs: c.ReadOnly,
		PidMode:        runconfig.PidMode(c.Pid),
		UTSMode:        runconfig.UTSMode(c.Uts),
		IpcMode:        runconfig.IpcMode(c.Ipc),
		PortBindings:   binding,
		RestartPolicy:  restart,
		SecurityOpt:    c.SecurityOpt,
	}

	return config, host_config, nil
}

func parseDevices(devices []string) ([]runconfig.DeviceMapping, error) {
	// parse device mappings
	deviceMappings := []runconfig.DeviceMapping{}
	for _, device := range devices {
		deviceMapping, err := runconfig.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, deviceMapping)
	}

	return deviceMappings, nil
}
