package config

import (
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/rancherio/rancher-compose/librcompose/project"
)

const (
	CONSOLE_CONTAINER  = "console"
	DOCKER_BIN         = "/usr/bin/docker"
	DOCKER_SYSTEM_HOME = "/var/lib/system-docker"
	DOCKER_SYSTEM_HOST = "unix:///var/run/system-docker.sock"
	DOCKER_HOST        = "unix:///var/run/docker.sock"
	IMAGES_PATH        = "/"
	IMAGES_PATTERN     = "images*.tar"
	SYS_INIT           = "/sbin/init-sys"
	USER_INIT          = "/sbin/init-user"
	MODULES_ARCHIVE    = "/modules.tar"
	DEBUG              = false

	LABEL         = "label"
	HASH          = "io.rancher.os.hash"
	ID            = "io.rancher.os.id"
	DETACH        = "io.rancher.os.detach"
	REMOVE        = "io.rancher.os.remove"
	CREATE_ONLY   = "io.rancher.os.createonly"
	RELOAD_CONFIG = "io.rancher.os.reloadconfig"
	SCOPE         = "io.rancher.os.scope"
	SYSTEM        = "system"

	OsConfigFile          = "/os-config.yml"
	CloudConfigFile       = "/var/lib/rancher/conf/cloud-config.yml"
	CloudConfigScriptFile = "/var/lib/rancher/conf/cloud-config-script"
	MetaDataFile          = "/var/lib/rancher/conf/metadata"
	LocalConfigFile       = "/var/lib/rancher/conf/cloud-config-local.yml"
	PrivateConfigFile     = "/var/lib/rancher/conf/cloud-config-private.yml"
)

var (
	VERSION string
)

type ContainerConfig struct {
	Id             string                 `yaml:"id,omitempty"`
	Cmd            string                 `yaml:"run,omitempty"`
	MigrateVolumes bool                   `yaml:"migrate_volumes,omitempty"`
	ReloadConfig   bool                   `yaml:"reload_config,omitempty"`
	CreateOnly     bool                   `yaml:create_only,omitempty`
	Service        *project.ServiceConfig `yaml:service,omitempty`
}

type Repository struct {
	Url string `yaml:url,omitempty`
}

type Repositories map[string]Repository

type CloudConfig struct {
	SSHAuthorizedKeys []string      `yaml:"ssh_authorized_keys"`
	WriteFiles        []config.File `yaml:"write_files"`
	Hostname          string        `yaml:"hostname"`
	Users             []config.User `yaml:"users"`

	Rancher RancherConfig `yaml:"rancher,omitempty"`
}

type RancherConfig struct {
	Environment         map[string]string                 `yaml:"environment,omitempty"`
	Services            map[string]*project.ServiceConfig `yaml:"services,omitempty"`
	BootstrapContainers map[string]*project.ServiceConfig `yaml:"bootstrap,omitempty"`
	Autoformat          map[string]*project.ServiceConfig `yaml:"autoformat,omitempty"`
	BootstrapDocker     DockerConfig                      `yaml:"bootstrap_docker,omitempty"`
	CloudInit           CloudInit                         `yaml:"cloud_init,omitempty"`
	Console             ConsoleConfig                     `yaml:"console,omitempty"`
	Debug               bool                              `yaml:"debug,omitempty"`
	Disable             []string                          `yaml:"disable,omitempty"`
	ServicesInclude     map[string]bool                   `yaml:"services_include,omitempty"`
	Modules             []string                          `yaml:"modules,omitempty"`
	Network             NetworkConfig                     `yaml:"network,omitempty"`
	Repositories        Repositories                      `yaml:"repositories,omitempty"`
	Ssh                 SshConfig                         `yaml:"ssh,omitempty"`
	State               StateConfig                       `yaml:"state,omitempty"`
	SystemDocker        DockerConfig                      `yaml:"system_docker,omitempty"`
	Upgrade             UpgradeConfig                     `yaml:"upgrade,omitempty"`
	UserContainers      []ContainerConfig                 `yaml:"user_containers,omitempty"`
	UserDocker          DockerConfig                      `yaml:"user_docker,omitempty"`
}

type ConsoleConfig struct {
	Tail       bool `yaml:"tail,omitempty"`
	Persistent bool `yaml:"persistent,omitempty"`
}

type UpgradeConfig struct {
	Url      string `yaml:"url,omitempty"`
	Image    string `yaml:"image,omitempty"`
	Rollback string `yaml:"rollback,omitempty"`
}

type DnsConfig struct {
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
	Domain      string   `yaml:"domain,omitempty"`
}

type NetworkConfig struct {
	Dns        DnsConfig                  `yaml:"dns,omitempty"`
	Interfaces map[string]InterfaceConfig `yaml:"interfaces,omitempty"`
	PostRun    *ContainerConfig           `yaml:"post_run,omitempty"`
}

type InterfaceConfig struct {
	Match   string `yaml:"match,omitempty"`
	DHCP    bool   `yaml:"dhcp,omitempty"`
	Address string `yaml:"address,omitempty"`
	IPV4LL  bool   `yaml:"ipv4ll,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
	MTU     int    `yaml:"mtu,omitempty"`
	Bridge  bool   `yaml:"bridge,omitempty"`
}

type DockerConfig struct {
	TLS        bool     `yaml:"tls,omitempty"`
	TLSArgs    []string `yaml:"tls_args,flow,omitempty"`
	Args       []string `yaml:"args,flow,omitempty"`
	ExtraArgs  []string `yaml:"extra_args,flow,omitempty"`
	ServerCert string   `yaml:"server_cert,omitempty"`
	ServerKey  string   `yaml:"server_key,omitempty"`
	CACert     string   `yaml:"ca_cert,omitempty"`
	CAKey      string   `yaml:"ca_key,omitempty"`
}

type SshConfig struct {
	Keys map[string]string `yaml:"keys,omitempty"`
}

type StateConfig struct {
	FsType     string   `yaml:"fstype,omitempty"`
	Dev        string   `yaml:"dev,omitempty"`
	Required   bool     `yaml:"required,omitempty"`
	Autoformat []string `yaml:"autoformat,omitempty"`
	FormatZero bool     `yaml:"formatzero,omitempty"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}

func init() {
	if VERSION == "" {
		VERSION = "v0.0.0-dev"
	}
}
