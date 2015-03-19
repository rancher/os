package config

const (
	VERSION            = "0.2.0-dev"
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
)

var (
	CloudConfigFile   = "/var/lib/rancher/conf/cloud-config-rancher.yml"
	ConfigFile        = "/var/lib/rancher/conf/rancher.yml"
	PrivateConfigFile = "/var/lib/rancher/conf/rancher-private.yml"
)

type ContainerConfig struct {
	Id             string `yaml:"id,omitempty"`
	Cmd            string `yaml:"run,omitempty"`
	MigrateVolumes bool   `yaml:"migrate_volumes,omitempty"`
	ReloadConfig   bool   `yaml:"reload_config,omitempty"`
}

type Config struct {
	Addons              map[string]Config `yaml:"addons,omitempty"`
	BootstrapContainers []ContainerConfig `yaml:"bootstrap_containers,omitempty"`
	CloudInit           CloudInit         `yaml:"cloud_init,omitempty"`
	Console             ConsoleConfig     `yaml:"console,omitempty"`
	Debug               bool              `yaml:"debug,omitempty"`
	Disable             []string          `yaml:"disable,omitempty"`
	EnabledAddons       []string          `yaml:"enabled_addons,omitempty"`
	Modules             []string          `yaml:"modules,omitempty"`
	Network             NetworkConfig     `yaml:"network,omitempty"`
	Ssh                 SshConfig         `yaml:"ssh,omitempty"`
	State               StateConfig       `yaml:"state,omitempty"`
	SystemContainers    []ContainerConfig `yaml:"system_containers,omitempty"`
	SystemDocker        DockerConfig      `yaml:"system_docker,omitempty"`
	Upgrade             UpgradeConfig     `yaml:"upgrade,omitempty"`
	UserContainers      []ContainerConfig `yaml:"user_containers,omitempty"`
	UserDocker          DockerConfig      `yaml:"user_docker,omitempty"`
}

type ConsoleConfig struct {
	Tail       bool `yaml:"tail,omitempty"`
	Persistent bool `yaml:"persistent,omitempty"`
}

type UpgradeConfig struct {
	Url string `yaml:"url,omitempty"`
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
	Gateway string `yaml:"gateway,omitempty"`
	MTU     int    `yaml:"mtu,omitempty"`
}

type DockerConfig struct {
	TLS        bool     `yaml:"tls,omitempty"`
	TLSArgs    []string `yaml:"tls_args,flow,omitempty"`
	Args       []string `yaml:"args,flow,omitempty"`
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
