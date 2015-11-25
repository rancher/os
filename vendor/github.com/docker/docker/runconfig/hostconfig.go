package runconfig

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/docker/docker/pkg/nat"
	"github.com/docker/docker/pkg/ulimit"
)

// KeyValuePair is a structure that hold a value for a key.
type KeyValuePair struct {
	Key   string
	Value string
}

// NetworkMode represents the container network stack.
type NetworkMode string

// IpcMode represents the container ipc stack.
type IpcMode string

// IsPrivate indicates whether the container uses it's private ipc stack.
func (n IpcMode) IsPrivate() bool {
	return !(n.IsHost() || n.IsContainer())
}

// IsHost indicates whether the container uses the host's ipc stack.
func (n IpcMode) IsHost() bool {
	return n == "host"
}

// IsContainer indicates whether the container uses a container's ipc stack.
func (n IpcMode) IsContainer() bool {
	parts := strings.SplitN(string(n), ":", 2)
	return len(parts) > 1 && parts[0] == "container"
}

// Valid indicates whether the ipc stack is valid.
func (n IpcMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	case "container":
		if len(parts) != 2 || parts[1] == "" {
			return false
		}
	default:
		return false
	}
	return true
}

// Container returns the name of the container ipc stack is going to be used.
func (n IpcMode) Container() string {
	parts := strings.SplitN(string(n), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// UTSMode represents the UTS namespace of the container.
type UTSMode string

// IsPrivate indicates whether the container uses it's private UTS namespace.
func (n UTSMode) IsPrivate() bool {
	return !(n.IsHost())
}

// IsHost indicates whether the container uses the host's UTS namespace.
func (n UTSMode) IsHost() bool {
	return n == "host"
}

// Valid indicates whether the UTS namespace is valid.
func (n UTSMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// PidMode represents the pid stack of the container.
type PidMode string

// IsPrivate indicates whether the container uses it's private pid stack.
func (n PidMode) IsPrivate() bool {
	return !(n.IsHost())
}

// IsHost indicates whether the container uses the host's pid stack.
func (n PidMode) IsHost() bool {
	return n == "host"
}

// Valid indicates whether the pid stack is valid.
func (n PidMode) Valid() bool {
	parts := strings.Split(string(n), ":")
	switch mode := parts[0]; mode {
	case "", "host":
	default:
		return false
	}
	return true
}

// DeviceMapping represents the device mapping between the host and the container.
type DeviceMapping struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}

// RestartPolicy represents the restart policies of the container.
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// IsNone indicates whether the container has the "no" restart policy.
// This means the container will not automatically restart when exiting.
func (rp *RestartPolicy) IsNone() bool {
	return rp.Name == "no"
}

// IsAlways indicates whether the container has the "always" restart policy.
// This means the container will automatically restart regardless of the exit status.
func (rp *RestartPolicy) IsAlways() bool {
	return rp.Name == "always"
}

// IsOnFailure indicates whether the container has the "on-failure" restart policy.
// This means the contain will automatically restart of exiting with a non-zero exit status.
func (rp *RestartPolicy) IsOnFailure() bool {
	return rp.Name == "on-failure"
}

// LogConfig represents the logging configuration of the container.
type LogConfig struct {
	Type   string
	Config map[string]string
}

// LxcConfig represents the specific LXC configuration of the container.
type LxcConfig struct {
	values []KeyValuePair
}

// MarshalJSON marshals (or serializes) the LxcConfig into JSON.
func (c *LxcConfig) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte{}, nil
	}
	return json.Marshal(c.Slice())
}

// UnmarshalJSON unmarshals (or deserializes) the specified byte slices from JSON to
// a LxcConfig.
func (c *LxcConfig) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	var kv []KeyValuePair
	if err := json.Unmarshal(b, &kv); err != nil {
		var h map[string]string
		if err := json.Unmarshal(b, &h); err != nil {
			return err
		}
		for k, v := range h {
			kv = append(kv, KeyValuePair{k, v})
		}
	}
	c.values = kv

	return nil
}

// Len returns the number of specific lxc configuration.
func (c *LxcConfig) Len() int {
	if c == nil {
		return 0
	}
	return len(c.values)
}

// Slice returns the specific lxc configuration into a slice of KeyValuePair.
func (c *LxcConfig) Slice() []KeyValuePair {
	if c == nil {
		return nil
	}
	return c.values
}

// NewLxcConfig creates a LxcConfig from the specified slice of KeyValuePair.
func NewLxcConfig(values []KeyValuePair) *LxcConfig {
	return &LxcConfig{values}
}

// CapList represents the list of capabilities of the container.
type CapList struct {
	caps []string
}

// MarshalJSON marshals (or serializes) the CapList into JSON.
func (c *CapList) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte{}, nil
	}
	return json.Marshal(c.Slice())
}

// UnmarshalJSON unmarshals (or deserializes) the specified byte slices
// from JSON to a CapList.
func (c *CapList) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	var caps []string
	if err := json.Unmarshal(b, &caps); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		caps = append(caps, s)
	}
	c.caps = caps

	return nil
}

// Len returns the number of specific kernel capabilities.
func (c *CapList) Len() int {
	if c == nil {
		return 0
	}
	return len(c.caps)
}

// Slice returns the specific capabilities into a slice of KeyValuePair.
func (c *CapList) Slice() []string {
	if c == nil {
		return nil
	}
	return c.caps
}

// NewCapList creates a CapList from a slice of string.
func NewCapList(caps []string) *CapList {
	return &CapList{caps}
}

// HostConfig the non-portable Config structure of a container.
// Here, "non-portable" means "dependent of the host we are running on".
// Portable information *should* appear in Config.
type HostConfig struct {
	Binds            []string         // List of volume bindings for this container
	ContainerIDFile  string           // File (path) where the containerId is written
	LxcConf          *LxcConfig       // Additional lxc configuration
	Memory           int64            // Memory limit (in bytes)
	MemorySwap       int64            // Total memory usage (memory + swap); set `-1` to disable swap
	CPUShares        int64            `json:"CpuShares"` // CPU shares (relative weight vs. other containers)
	CPUPeriod        int64            `json:"CpuPeriod"` // CPU CFS (Completely Fair Scheduler) period
	CpusetCpus       string           // CpusetCpus 0-2, 0,1
	CpusetMems       string           // CpusetMems 0-2, 0,1
	CPUQuota         int64            `json:"CpuQuota"` // CPU CFS (Completely Fair Scheduler) quota
	BlkioWeight      int64            // Block IO weight (relative weight vs. other containers)
	OomKillDisable   bool             // Whether to disable OOM Killer or not
	MemorySwappiness *int64           // Tuning container memory swappiness behaviour
	Privileged       bool             // Is the container in privileged mode
	PortBindings     nat.PortMap      // Port mapping between the exposed port (container) and the host
	Links            []string         // List of links (in the name:alias form)
	PublishAllPorts  bool             // Should docker publish all exposed port for the container
	DNS              []string         `json:"Dns"`       // List of DNS server to lookup
	DNSSearch        []string         `json:"DnsSearch"` // List of DNSSearch to look for
	ExtraHosts       []string         // List of extra hosts
	VolumesFrom      []string         // List of volumes to take from other container
	Devices          []DeviceMapping  // List of devices to map inside the container
	NetworkMode      NetworkMode      // Network namespace to use for the container
	IpcMode          IpcMode          // IPC namespace to use for the container
	PidMode          PidMode          // PID namespace to use for the container
	UTSMode          UTSMode          // UTS namespace to use for the container
	CapAdd           *CapList         // List of kernel capabilities to add to the container
	CapDrop          *CapList         // List of kernel capabilities to remove from the container
	GroupAdd         []string         // List of additional groups that the container process will run as
	RestartPolicy    RestartPolicy    // Restart policy to be used for the container
	SecurityOpt      []string         // List of string values to customize labels for MLS systems, such as SELinux.
	ReadonlyRootfs   bool             // Is the container root filesystem in read-only
	Ulimits          []*ulimit.Ulimit // List of ulimits to be set in the container
	LogConfig        LogConfig        // Configuration of the logs for this container
	CgroupParent     string           // Parent cgroup.
	ConsoleSize      [2]int           // Initial console size on Windows
}

// MergeConfigs merges the specified container Config and HostConfig.
// It creates a ContainerConfigWrapper.
func MergeConfigs(config *Config, hostConfig *HostConfig) *ContainerConfigWrapper {
	return &ContainerConfigWrapper{
		config,
		hostConfig,
		"", nil,
	}
}

// DecodeHostConfig creates a HostConfig based on the specified Reader.
// It assumes the content of the reader will be JSON, and decodes it.
func DecodeHostConfig(src io.Reader) (*HostConfig, error) {
	decoder := json.NewDecoder(src)

	var w ContainerConfigWrapper
	if err := decoder.Decode(&w); err != nil {
		return nil, err
	}

	hc := w.GetHostConfig()

	return hc, nil
}
