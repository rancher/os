package project

type Event string

const (
	CONTAINER_ID = "container_id"

	CONTAINER_CREATED = Event("Created container")
	CONTAINER_STARTED = Event("Started container")

	SERVICE_ADD           = Event("Adding")
	SERVICE_UP_START      = Event("Starting")
	SERVICE_UP_IGNORED    = Event("Ignoring")
	SERVICE_UP            = Event("Started")
	SERVICE_CREATE_START  = Event("Creating")
	SERVICE_CREATE        = Event("Created")
	SERVICE_DELETE_START  = Event("Deleting")
	SERVICE_DELETE        = Event("Deleted")
	SERVICE_DOWN_START    = Event("Stopping")
	SERVICE_DOWN          = Event("Stopped")
	SERVICE_RESTART_START = Event("Restarting")
	SERVICE_RESTART       = Event("Restarted")
	SERVICE_PULL_START    = Event("Pulling")
	SERVICE_PULL          = Event("Pulled")
	SERVICE_KILL_START    = Event("Killing")
	SERVICE_KILL          = Event("Killed")
	SERVICE_START_START   = Event("Starting")
	SERVICE_START         = Event("Started")
	SERVICE_BUILD_START   = Event("Building")
	SERVICE_BUILD         = Event("Built")

	PROJECT_DOWN_START     = Event("Stopping project")
	PROJECT_DOWN_DONE      = Event("Project stopped")
	PROJECT_CREATE_START   = Event("Creating project")
	PROJECT_CREATE_DONE    = Event("Project created")
	PROJECT_UP_START       = Event("Starting project")
	PROJECT_UP_DONE        = Event("Project started")
	PROJECT_DELETE_START   = Event("Deleting project")
	PROJECT_DELETE_DONE    = Event("Project deleted")
	PROJECT_RESTART_START  = Event("Restarting project")
	PROJECT_RESTART_DONE   = Event("Project restarted")
	PROJECT_RELOAD         = Event("Reloading project")
	PROJECT_RELOAD_TRIGGER = Event("Triggering project reload")
	PROJECT_KILL_START     = Event("Killing project")
	PROJECT_KILL_DONE      = Event("Project killed")
	PROJECT_START_START    = Event("Starting project")
	PROJECT_START_DONE     = Event("Project started")
	PROJECT_BUILD_START    = Event("Building project")
	PROJECT_BUILD_DONE     = Event("Project built")
)

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
