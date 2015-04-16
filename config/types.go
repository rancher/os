package config

import "github.com/rancherio/rancher-compose/project"

const (
	DEFAULT_IMAGE_VERSION = "v0.3.0-rc2"
	CONSOLE_CONTAINER     = "console"
	DOCKER_BIN            = "/usr/bin/docker"
	DOCKER_SYSTEM_HOME    = "/var/lib/system-docker"
	DOCKER_SYSTEM_HOST    = "unix:///var/run/system-docker.sock"
	DOCKER_HOST           = "unix:///var/run/docker.sock"
	IMAGES_PATH           = "/"
	IMAGES_PATTERN        = "images*.tar"
	SYS_INIT              = "/sbin/init-sys"
	USER_INIT             = "/sbin/init-user"
	MODULES_ARCHIVE       = "/modules.tar"
	DEBUG                 = false

	LABEL         = "label"
	HASH          = "io.rancher.os.hash"
	ID            = "io.rancher.os.id"
	DETACH        = "io.rancher.os.detach"
	REMOVE        = "io.rancher.os.remove"
	CREATE_ONLY   = "io.rancher.os.createonly"
	RELOAD_CONFIG = "io.rancher.os.reloadconfig"
)

var (
	VERSION           string
	IMAGE_VERSION     string
	CloudConfigFile   = "/var/lib/rancher/conf/cloud-config-rancher.yml"
	ConfigFile        = "/var/lib/rancher/conf/rancher.yml"
	PrivateConfigFile = "/var/lib/rancher/conf/rancher-private.yml"
)

type ContainerConfig struct {
	Id             string                 `yaml:"id,omitempty"`
	Cmd            string                 `yaml:"run,omitempty"`
	MigrateVolumes bool                   `yaml:"migrate_volumes,omitempty"`
	ReloadConfig   bool                   `yaml:"reload_config,omitempty"`
	CreateOnly     bool                   `yaml:create_only,omitempty`
	Service        *project.ServiceConfig `yaml:service,omitempty`
}

type Config struct {
	Environment         map[string]string                 `yaml:"environment,omitempty"`
	BundledServices     map[string]Config                 `yaml:"bundled_services,omitempty"`
	BootstrapContainers map[string]*project.ServiceConfig `yaml:"bootstrap_containers,omitempty"`
	BootstrapDocker     DockerConfig                      `yaml:"bootstrap_docker,omitempty"`
	CloudInit           CloudInit                         `yaml:"cloud_init,omitempty"`
	Console             ConsoleConfig                     `yaml:"console,omitempty"`
	Debug               bool                              `yaml:"debug,omitempty"`
	Disable             []string                          `yaml:"disable,omitempty"`
	Services            map[string]bool                   `yaml:"services,omitempty"`
	Modules             []string                          `yaml:"modules,omitempty"`
	Network             NetworkConfig                     `yaml:"network,omitempty"`
	Ssh                 SshConfig                         `yaml:"ssh,omitempty"`
	State               StateConfig                       `yaml:"state,omitempty"`
	SystemContainers    map[string]*project.ServiceConfig `yaml:"system_containers,omitempty"`
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
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}

func init() {
	if VERSION == "" {
		VERSION = "v0.0.0-dev"
	}
	if IMAGE_VERSION == "" {
		IMAGE_VERSION = DEFAULT_IMAGE_VERSION
	}
}
