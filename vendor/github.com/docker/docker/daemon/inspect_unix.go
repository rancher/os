// +build !windows

package daemon

import "github.com/docker/docker/api/types"

// This sets platform-specific fields
func setPlatformSpecificContainerFields(container *Container, contJSONBase *types.ContainerJSONBase) *types.ContainerJSONBase {
	contJSONBase.AppArmorProfile = container.AppArmorProfile
	contJSONBase.ResolvConfPath = container.ResolvConfPath
	contJSONBase.HostnamePath = container.HostnamePath
	contJSONBase.HostsPath = container.HostsPath

	return contJSONBase
}

func (daemon *Daemon) ContainerInspectPre120(name string) (*types.ContainerJSONPre120, error) {
	container, err := daemon.Get(name)
	if err != nil {
		return nil, err
	}

	container.Lock()
	defer container.Unlock()

	base, err := daemon.getInspectData(container)
	if err != nil {
		return nil, err
	}

	volumes := make(map[string]string)
	volumesRW := make(map[string]bool)
	for _, m := range container.MountPoints {
		volumes[m.Destination] = m.Path()
		volumesRW[m.Destination] = m.RW
	}

	config := &types.ContainerConfig{
		container.Config,
		container.hostConfig.Memory,
		container.hostConfig.MemorySwap,
		container.hostConfig.CPUShares,
		container.hostConfig.CpusetCpus,
	}

	return &types.ContainerJSONPre120{base, volumes, volumesRW, config}, nil
}

func addMountPoints(container *Container) []types.MountPoint {
	mountPoints := make([]types.MountPoint, 0, len(container.MountPoints))
	for _, m := range container.MountPoints {
		mountPoints = append(mountPoints, types.MountPoint{
			Name:        m.Name,
			Source:      m.Path(),
			Destination: m.Destination,
			Driver:      m.Driver,
			Mode:        m.Mode,
			RW:          m.RW,
		})
	}
	return mountPoints
}
