package project

import "fmt"

type Event int

const (
	CONTAINER_ID = "container_id"

	NO_EVENT = Event(iota)

	CONTAINER_CREATED = Event(iota)
	CONTAINER_STARTED = Event(iota)

	SERVICE_ADD           = Event(iota)
	SERVICE_UP_START      = Event(iota)
	SERVICE_UP_IGNORED    = Event(iota)
	SERVICE_UP            = Event(iota)
	SERVICE_CREATE_START  = Event(iota)
	SERVICE_CREATE        = Event(iota)
	SERVICE_DELETE_START  = Event(iota)
	SERVICE_DELETE        = Event(iota)
	SERVICE_DOWN_START    = Event(iota)
	SERVICE_DOWN          = Event(iota)
	SERVICE_RESTART_START = Event(iota)
	SERVICE_RESTART       = Event(iota)
	SERVICE_PULL_START    = Event(iota)
	SERVICE_PULL          = Event(iota)
	SERVICE_KILL_START    = Event(iota)
	SERVICE_KILL          = Event(iota)
	SERVICE_START_START   = Event(iota)
	SERVICE_START         = Event(iota)
	SERVICE_BUILD_START   = Event(iota)
	SERVICE_BUILD         = Event(iota)

	PROJECT_DOWN_START     = Event(iota)
	PROJECT_DOWN_DONE      = Event(iota)
	PROJECT_CREATE_START   = Event(iota)
	PROJECT_CREATE_DONE    = Event(iota)
	PROJECT_UP_START       = Event(iota)
	PROJECT_UP_DONE        = Event(iota)
	PROJECT_DELETE_START   = Event(iota)
	PROJECT_DELETE_DONE    = Event(iota)
	PROJECT_RESTART_START  = Event(iota)
	PROJECT_RESTART_DONE   = Event(iota)
	PROJECT_RELOAD         = Event(iota)
	PROJECT_RELOAD_TRIGGER = Event(iota)
	PROJECT_KILL_START     = Event(iota)
	PROJECT_KILL_DONE      = Event(iota)
	PROJECT_START_START    = Event(iota)
	PROJECT_START_DONE     = Event(iota)
	PROJECT_BUILD_START    = Event(iota)
	PROJECT_BUILD_DONE     = Event(iota)
)

func (e Event) String() string {
	var m string
	switch e {
	case CONTAINER_CREATED:
		m = "Created container"
	case CONTAINER_STARTED:
		m = "Started container"

	case SERVICE_ADD:
		m = "Adding"
	case SERVICE_UP_START:
		m = "Starting"
	case SERVICE_UP_IGNORED:
		m = "Ignoring"
	case SERVICE_UP:
		m = "Started"
	case SERVICE_CREATE_START:
		m = "Creating"
	case SERVICE_CREATE:
		m = "Created"
	case SERVICE_DELETE_START:
		m = "Deleting"
	case SERVICE_DELETE:
		m = "Deleted"
	case SERVICE_DOWN_START:
		m = "Stopping"
	case SERVICE_DOWN:
		m = "Stopped"
	case SERVICE_RESTART_START:
		m = "Restarting"
	case SERVICE_RESTART:
		m = "Restarted"
	case SERVICE_PULL_START:
		m = "Pulling"
	case SERVICE_PULL:
		m = "Pulled"
	case SERVICE_KILL_START:
		m = "Killing"
	case SERVICE_KILL:
		m = "Killed"
	case SERVICE_START_START:
		m = "Starting"
	case SERVICE_START:
		m = "Started"
	case SERVICE_BUILD_START:
		m = "Building"
	case SERVICE_BUILD:
		m = "Built"

	case PROJECT_DOWN_START:
		m = "Stopping project"
	case PROJECT_DOWN_DONE:
		m = "Project stopped"
	case PROJECT_CREATE_START:
		m = "Creating project"
	case PROJECT_CREATE_DONE:
		m = "Project created"
	case PROJECT_UP_START:
		m = "Starting project"
	case PROJECT_UP_DONE:
		m = "Project started"
	case PROJECT_DELETE_START:
		m = "Deleting project"
	case PROJECT_DELETE_DONE:
		m = "Project deleted"
	case PROJECT_RESTART_START:
		m = "Restarting project"
	case PROJECT_RESTART_DONE:
		m = "Project restarted"
	case PROJECT_RELOAD:
		m = "Reloading project"
	case PROJECT_RELOAD_TRIGGER:
		m = "Triggering project reload"
	case PROJECT_KILL_START:
		m = "Killing project"
	case PROJECT_KILL_DONE:
		m = "Project killed"
	case PROJECT_START_START:
		m = "Starting project"
	case PROJECT_START_DONE:
		m = "Project started"
	case PROJECT_BUILD_START:
		m = "Building project"
	case PROJECT_BUILD_DONE:
		m = "Project built"
	}

	if m == "" {
		m = fmt.Sprintf("Event: %d", int(e))
	}

	return m
}

type InfoPart struct {
	Key, Value string
}

type InfoSet []Info
type Info []InfoPart

type ServiceConfig struct {
	Build         string            `yaml:"build,omitempty"`
	CapAdd        []string          `yaml:"cap_add,omitempty"`
	CapDrop       []string          `yaml:"cap_drop,omitempty"`
	CpuSet        string            `yaml:"cpu_set,omitempty"`
	CpuShares     int64             `yaml:"cpu_shares,omitempty"`
	Command       Command           `yaml:"command"` // omitempty breaks serialization!
	Detach        string            `yaml:"detach,omitempty"`
	Devices       []string          `yaml:"devices,omitempty"`
	Dns           Stringorslice     `yaml:"dns"`        // omitempty breaks serialization!
	DnsSearch     Stringorslice     `yaml:"dns_search"` // omitempty breaks serialization!
	Dockerfile    string            `yaml:"dockerfile,omitempty"`
	DomainName    string            `yaml:"domainname,omitempty"`
	Entrypoint    Command           `yaml:"entrypoint"`  // omitempty breaks serialization!
	EnvFile       Stringorslice     `yaml:"env_file"`    // omitempty breaks serialization!
	Environment   MaporEqualSlice   `yaml:"environment"` // omitempty breaks serialization!
	Hostname      string            `yaml:"hostname,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Labels        SliceorMap        `yaml:"labels"` // omitempty breaks serialization!
	Links         MaporColonSlice   `yaml:"links"`  // omitempty breaks serialization!
	LogDriver     string            `yaml:"log_driver,omitempty"`
	MemLimit      int64             `yaml:"mem_limit,omitempty"`
	MemSwapLimit  int64             `yaml:"mem_swap_limit,omitempty"`
	Name          string            `yaml:"name,omitempty"`
	Net           string            `yaml:"net,omitempty"`
	Pid           string            `yaml:"pid,omitempty"`
	Uts           string            `yaml:"uts,omitempty"`
	Ipc           string            `yaml:"ipc,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Privileged    bool              `yaml:"privileged,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	ReadOnly      bool              `yaml:"read_only,omitempty"`
	StdinOpen     bool              `yaml:"stdin_open,omitempty"`
	SecurityOpt   []string          `yaml:"security_opt,omitempty"`
	Tty           bool              `yaml:"tty,omitempty"`
	User          string            `yaml:"user,omitempty"`
	VolumeDriver  string            `yaml:"volume_driver,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	VolumesFrom   []string          `yaml:"volumes_from,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	Expose        []string          `yaml:"expose,omitempty"`
	ExternalLinks []string          `yaml:"external_links,omitempty"`
	LogOpt        map[string]string `yaml:"log_opt,omitempty"`
	ExtraHosts    []string          `yaml:"extra_hosts,omitempty"`
}

type EnvironmentLookup interface {
	Lookup(key, serviceName string, config *ServiceConfig) []string
}

type ConfigLookup interface {
	Lookup(file, relativeTo string) ([]byte, string, error)
}

type Project struct {
	Name           string
	Configs        map[string]*ServiceConfig
	File           string
	ReloadCallback func() error
	context        *Context
	reload         []string
	upCount        int
	listeners      []chan<- ProjectEvent
	hasListeners   bool
}

type Service interface {
	Info() (InfoSet, error)
	Name() string
	Build() error
	Create() error
	Up() error
	Start() error
	Down() error
	Delete() error
	Restart() error
	Log() error
	Pull() error
	Kill() error
	Config() *ServiceConfig
	DependentServices() []ServiceRelationship
	Containers() ([]Container, error)
	Scale(count int) error
}

type Container interface {
	Id() (string, error)
	Name() string
	Port(port string) (string, error)
}

type ServiceFactory interface {
	Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error)
}

type ServiceRelationshipType string

const REL_TYPE_LINK = ServiceRelationshipType("")
const REL_TYPE_NET_NAMESPACE = ServiceRelationshipType("netns")
const REL_TYPE_IPC_NAMESPACE = ServiceRelationshipType("ipc")
const REL_TYPE_VOLUMES_FROM = ServiceRelationshipType("volumesFrom")

type ServiceRelationship struct {
	Target, Alias string
	Type          ServiceRelationshipType
	Optional      bool
}

func NewServiceRelationship(nameAlias string, relType ServiceRelationshipType) ServiceRelationship {
	name, alias := NameAlias(nameAlias)
	return ServiceRelationship{
		Target: name,
		Alias:  alias,
		Type:   relType,
	}
}
