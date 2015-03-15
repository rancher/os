package config

const (
	VERSION            = "0.0.1"
	CONSOLE_CONTAINER  = "console"
	DOCKER_BIN         = "/usr/bin/docker"
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
	Debug   bool     `yaml:"debug,omitempty"`
	Disable []string `yaml:"disable,omitempty"`
	Dns     []string `yaml:"dns,flow,omitempty"`
	//Rescue              bool              `yaml:"rescue,omitempty"`
	//RescueContainer     *ContainerConfig  `yaml:"rescue_container,omitempty"`
	Console             ConsoleConfig     `yaml:"console,omitempty"`
	State               ConfigState       `yaml:"state,omitempty"`
	Userdocker          UserDockerConfig  `yaml:"userdocker,omitempty"`
	Upgrade             UpgradeConfig     `yaml:"upgrade,omitempty"`
	BootstrapContainers []ContainerConfig `yaml:"bootstrap_containers,omitempty"`
	SystemContainers    []ContainerConfig `yaml:"system_containers,omitempty"`
	UserContainers      []ContainerConfig `yaml:"user_containers,omitempty"`
	SystemDockerArgs    []string          `yaml:"system_docker_args,flow,omitempty"`
	Modules             []string          `yaml:"modules,omitempty"`
	CloudInit           CloudInit         `yaml:"cloud_init,omitempty"`
	Ssh                 SshConfig         `yaml:"ssh,omitempty"`
	EnabledAddons       []string          `yaml:"enabled_addons,omitempty"`
	Addons              map[string]Config `yaml:"addons,omitempty"`
	Network             NetworkConfig     `yaml:"network,omitempty"`
}

type ConsoleConfig struct {
	Tail      bool `yaml:"tail,omitempty"`
	Ephemeral bool `yaml:"ephemeral,omitempty"`
}

type UpgradeConfig struct {
	Url string `yaml:"url,omitempty"`
}

type NetworkConfig struct {
	Interfaces []InterfaceConfig `yaml:"interfaces,omitempty"`
	PostRun    *ContainerConfig  `yaml:"post_run,omitempty"`
}

type InterfaceConfig struct {
	Match   string `yaml:"match,omitempty"`
	DHCP    bool   `yaml:"dhcp,omitempty"`
	Address string `yaml:"address,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
	MTU     int    `yaml:"mtu,omitempty"`
}

type UserDockerConfig struct {
	UseTLS        bool   `yaml:"use_tls,omitempty"`
	TLSServerCert string `yaml:"tls_server_cert,omitempty"`
	TLSServerKey  string `yaml:"tls_server_key,omitempty"`
	TLSCACert     string `yaml:"tls_ca_cert,omitempty"`
}

type SshConfig struct {
	Keys map[string]string `yaml:"keys,omitempty"`
}

type ConfigState struct {
	FsType   string `yaml:"fstype,omitempty"`
	Dev      string `yaml:"dev,omitempty"`
	Required bool   `yaml:"required,omitempty"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}
