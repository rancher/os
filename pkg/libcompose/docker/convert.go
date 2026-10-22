package docker

import (
	"fmt"
	"strings"

	"github.com/burmilla/os/pkg/libcompose/config"
	"github.com/burmilla/os/pkg/libcompose/project"
	"github.com/burmilla/os/pkg/libcompose/utils"
	"github.com/docker/docker/runconfig/opts"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/engine-api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

// ConfigWrapper wraps Config, HostConfig and NetworkingConfig for a container.
type ConfigWrapper struct {
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

// Filter filters the specified string slice with the specified function.
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

// ConvertToAPI converts a service configuration to a docker API container configuration.
func ConvertToAPI(s *Service) (*ConfigWrapper, error) {
	cfg, hostConfig, err := Convert(s.serviceConfig, s.context.Context)
	if err != nil {
		return nil, err
	}

	result := ConfigWrapper{
		Config:     cfg,
		HostConfig: hostConfig,
	}
	return &result, nil
}

func volumes(c *config.ServiceConfig, ctx project.Context) map[string]struct{} {
	volumeStrs := config.VolumesToStringSlice(c.Volumes)
	result := make(map[string]struct{}, len(volumeStrs))
	for _, v := range volumeStrs {
		if isVolume(v) {
			result[v] = struct{}{}
		}
	}
	return result
}

func restartPolicy(c *config.ServiceConfig) (*container.RestartPolicy, error) {
	restart, err := opts.ParseRestartPolicy(c.Restart)
	if err != nil {
		return nil, err
	}
	return &container.RestartPolicy{Name: restart.Name, MaximumRetryCount: restart.MaximumRetryCount}, nil
}

func ports(c *config.ServiceConfig) (map[nat.Port]struct{}, nat.PortMap, error) {
	portStrs := config.PortsToStringSlice(c.Ports)
	exposeStrs := make([]string, len(c.Expose))
	for i, e := range c.Expose {
		exposeStrs[i] = fmt.Sprint(e)
	}

	portSpecs, binding, err := nat.ParsePortSpecs(portStrs)
	if err != nil {
		return nil, nil, err
	}

	exPorts, _, err := nat.ParsePortSpecs(exposeStrs)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range exPorts {
		portSpecs[k] = v
	}

	exposedPorts := map[nat.Port]struct{}{}
	for k, v := range portSpecs {
		exposedPorts[nat.Port(k)] = v
	}

	portBindings := nat.PortMap{}
	for k, bv := range binding {
		dcbs := make([]nat.PortBinding, len(bv))
		for k, v := range bv {
			dcbs[k] = nat.PortBinding{HostIP: v.HostIP, HostPort: v.HostPort}
		}
		portBindings[nat.Port(k)] = dcbs
	}
	return exposedPorts, portBindings, nil
}

// Convert converts a service configuration to docker API structures (Config and HostConfig)
func Convert(c *config.ServiceConfig, ctx project.Context) (*container.Config, *container.HostConfig, error) {
	restartPolicy, err := restartPolicy(c)
	if err != nil {
		return nil, nil, err
	}

	exposedPorts, portBindings, err := ports(c)
	if err != nil {
		return nil, nil, err
	}

	deviceMappings, err := parseDevices(c.Devices)
	if err != nil {
		return nil, nil, err
	}

	var volumesFrom []string
	if c.VolumesFrom != nil {
		volumesFrom, err = getVolumesFrom(c.VolumesFrom, ctx.Project.ServiceConfigs, ctx.ProjectName)
		if err != nil {
			return nil, nil, err
		}
	}

	envSlice := config.EnvironmentToSlice(c.Environment)
	volumeStrs := config.VolumesToStringSlice(c.Volumes)

	containerConfig := &container.Config{
		Entrypoint:   strslice.StrSlice(utils.CopySlice([]string(c.Entrypoint))),
		Hostname:     c.Hostname,
		Domainname:   c.DomainName,
		User:         c.User,
		Env:          utils.CopySlice(envSlice),
		Cmd:          strslice.StrSlice(utils.CopySlice([]string(c.Command))),
		Image:        c.Image,
		Labels:       utils.CopyMap(c.Labels),
		ExposedPorts: exposedPorts,
		Tty:          c.Tty,
		OpenStdin:    c.StdinOpen,
		WorkingDir:   c.WorkingDir,
		Volumes:      volumes(c, ctx),
		MacAddress:   c.MacAddress,
	}

	ulimits := []*units.Ulimit{}
	if c.Ulimits != nil {
		for name, ulimit := range c.Ulimits {
			ulimits = append(ulimits, &units.Ulimit{
				Name: name,
				Soft: int64(ulimit.Soft),
				Hard: int64(ulimit.Hard),
			})
		}
	}

	logDriver := config.LoggingDriver(c)
	logOpts := config.LoggingOptions(c)

	networkMode := c.NetworkMode
	if networkMode == "" {
		networkMode = c.Net
	}

	resources := container.Resources{
		CgroupParent: c.CgroupParent,
		Memory:       int64(c.MemLimit),
		MemorySwap:   int64(c.MemSwapLimit),
		CPUShares:    c.CPUShares,
		CPUQuota:     c.CPUQuota,
		CpusetCpus:   c.CPUSet,
		Ulimits:      ulimits,
		Devices:      deviceMappings,
	}

	hostConfig := &container.HostConfig{
		VolumesFrom: volumesFrom,
		CapAdd:      strslice.StrSlice(utils.CopySlice(c.CapAdd)),
		CapDrop:     strslice.StrSlice(utils.CopySlice(c.CapDrop)),
		ExtraHosts:  utils.CopySlice(c.ExtraHosts.AsList()),
		Privileged:  c.Privileged,
		Binds:       Filter(volumeStrs, isBind),
		DNS:         utils.CopySlice([]string(c.DNS)),
		DNSSearch:   utils.CopySlice([]string(c.DNSSearch)),
		LogConfig: container.LogConfig{
			Type:   logDriver,
			Config: utils.CopyMap(logOpts),
		},
		NetworkMode:    container.NetworkMode(networkMode),
		ReadonlyRootfs: c.ReadOnly,
		OomScoreAdj:    int(c.OomScoreAdj),
		PidMode:        container.PidMode(c.Pid),
		UTSMode:        container.UTSMode(c.Uts),
		IpcMode:        container.IpcMode(c.Ipc),
		PortBindings:   portBindings,
		RestartPolicy:  *restartPolicy,
		SecurityOpt:    utils.CopySlice(c.SecurityOpt),
		VolumeDriver:   c.VolumeDriver,
		Resources:      resources,
	}

	return containerConfig, hostConfig, nil
}

func getVolumesFrom(volumesFrom []string, serviceConfigs *config.ServiceConfigs, projectName string) ([]string, error) {
	result := []string{}
	for _, volumeFrom := range volumesFrom {
		if serviceConfig, ok := serviceConfigs.Get(volumeFrom); ok {
			// It's a service - Use the first one
			name := fmt.Sprintf("%s_%s_1", projectName, volumeFrom)
			// If a container name is specified, use that instead
			if serviceConfig.ContainerName != "" {
				name = serviceConfig.ContainerName
			}
			result = append(result, name)
		} else {
			result = append(result, volumeFrom)
		}
	}
	return result, nil
}

func parseDevices(devices []string) ([]container.DeviceMapping, error) {
	deviceMappings := []container.DeviceMapping{}
	for _, device := range devices {
		v, err := opts.ParseDevice(device)
		if err != nil {
			return nil, err
		}
		deviceMappings = append(deviceMappings, container.DeviceMapping{
			PathOnHost:        v.PathOnHost,
			PathInContainer:   v.PathInContainer,
			CgroupPermissions: v.CgroupPermissions,
		})
	}
	return deviceMappings, nil
}
