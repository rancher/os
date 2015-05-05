package project

import (
	"strings"

	"github.com/rancherio/go-rancher/client"
	"gopkg.in/yaml.v2"
)

type Event string

const (
	CONTAINER_ID = "container_id"

	CONTAINER_STARTING = Event("Starting container")
	CONTAINER_CREATED  = Event("Created container")
	CONTAINER_STARTED  = Event("Started container")

	SERVICE_ADD      = Event("Adding")
	SERVICE_UP_START = Event("Starting")
	SERVICE_UP       = Event("Started")

	PROJECT_UP_START       = Event("Starting project")
	PROJECT_UP_DONE        = Event("Project started")
	PROJECT_RELOAD         = Event("Reloading project")
	PROJECT_RELOAD_TRIGGER = Event("Triggering project reload")
)

type Stringorslice struct {
	parts []string
}

func (s *Stringorslice) MarshalYAML() (interface{}, error) {
	if s == nil {
		return nil, nil
	}
	bytes, err := yaml.Marshal(s.Slice())
	return string(bytes), err
}

func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}
	return err
}

func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewStringorslice(parts ...string) *Stringorslice {
	return &Stringorslice{parts}
}

type SliceorMap struct {
	parts map[string]string
}

func (s *SliceorMap) MarshalYAML() (interface{}, error) {
	if s == nil {
		return nil, nil
	}
	bytes, err := yaml.Marshal(s.MapParts())
	return string(bytes), err
}

func (s *SliceorMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	mapType := make(map[string]string)
	err := unmarshal(&mapType)
	if err == nil {
		s.parts = mapType
		return nil
	}

	var sliceType []string
	var keyValueSlice []string
	var key string
	var value string

	err = unmarshal(&sliceType)
	if err == nil {
		mapType = make(map[string]string)
		for _, slice := range sliceType {
			keyValueSlice = strings.Split(slice, "=") //split up key and value into []string
			key = keyValueSlice[0]
			value = keyValueSlice[1]
			mapType[key] = value
		}
		s.parts = mapType
		return nil
	}
	return err
}

func (s *SliceorMap) MapParts() map[string]string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewSliceorMap(parts map[string]string) *SliceorMap {
	return &SliceorMap{parts}
}

type ServiceConfig struct {
	CapAdd      []string       `yaml:"cap_add,omitempty"`
	CapDrop     []string       `yaml:"cap_drop,omitempty"`
	CpuShares   int64          `yaml:"cpu_shares,omitempty"`
	Command     string         `yaml:"command,omitempty"`
	Detach      string         `yaml:"detach,omitempty"`
	Dns         *Stringorslice `yaml:"dns,omitempty"`
	DnsSearch   *Stringorslice `yaml:"dns_search,omitempty"`
	DomainName  string         `yaml:"domainname,omitempty"`
	Entrypoint  string         `yaml:"entrypoint,omitempty"`
	EnvFile     string         `yaml:"env_file,omitempty"`
	Environment []string       `yaml:"environment,omitempty"`
	Hostname    string         `yaml:"hostname,omitempty"`
	Image       string         `yaml:"image,omitempty"`
	Labels      *SliceorMap    `yaml:"labels,omitempty"`
	Links       []string       `yaml:"links,omitempty"`
	LogDriver   string         `yaml:"log_driver,omitempty"`
	MemLimit    int64          `yaml:"mem_limit,omitempty"`
	Name        string         `yaml:"name,omitempty"`
	Net         string         `yaml:"net,omitempty"`
	Pid         string         `yaml:"pid,omitempty"`
	Ipc         string         `yaml:"ipc,omitempty"`
	Ports       []string       `yaml:"ports,omitempty"`
	Privileged  bool           `yaml:"privileged,omitempty"`
	Restart     string         `yaml:"restart,omitempty"`
	ReadOnly    bool           `yaml:"read_only,omitempty"`
	StdinOpen   bool           `yaml:"stdin_open,omitempty"`
	Tty         bool           `yaml:"tty,omitempty"`
	User        string         `yaml:"user,omitempty"`
	Volumes     []string       `yaml:"volumes,omitempty"`
	VolumesFrom []string       `yaml:"volumes_from,omitempty"`
	WorkingDir  string         `yaml:"working_dir,omitempty"`
	//`yaml:"build,omitempty"`
	Expose        []string `yaml:"expose,omitempty"`
	ExternalLinks []string `yaml:"external_links,omitempty"`
}

type EnvironmentLookup interface {
	Lookup(key, serviceName string, config *ServiceConfig) []string
}

type Project struct {
	EnvironmentLookup EnvironmentLookup
	Name              string
	configs           map[string]*ServiceConfig
	reload            []string
	file              string
	content           []byte
	client            *client.RancherClient
	factory           ServiceFactory
	ReloadCallback    func() error
	upCount           int
	listeners         []chan<- ProjectEvent
}

type Service interface {
	Name() string
	Up() error
	Config() *ServiceConfig
}

type ServiceFactory interface {
	Create(project *Project, name string, serviceConfig *ServiceConfig) (Service, error)
}
