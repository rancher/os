package config

import (
	"fmt"
	"runtime"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/docker/engine-api/types"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config/yaml"
)

const (
	OEM                = "/usr/share/ros/oem"
	DOCKER_BIN         = "/usr/bin/docker"
	DOCKER_DIST_BIN    = "/usr/bin/docker.dist"
	ROS_BIN            = "/usr/bin/ros"
	SYSINIT_BIN        = "/usr/bin/ros-sysinit"
	DOCKER_SYSTEM_HOME = "/var/lib/system-docker"
	DOCKER_SYSTEM_HOST = "unix:///var/run/system-docker.sock"
	DOCKER_HOST        = "unix:///var/run/docker.sock"
	IMAGES_PATH        = "/usr/share/ros"
	IMAGES_PATTERN     = "images*.tar"
	MODULES_ARCHIVE    = "/modules.tar"
	DEBUG              = false
	SYSTEM_DOCKER_LOG  = "/var/log/system-docker.log"
	SYSTEM_DOCKER_BIN  = "/usr/bin/system-docker"

	LABEL         = "label"
	HASH          = "io.rancher.os.hash"
	ID            = "io.rancher.os.id"
	DETACH        = "io.rancher.os.detach"
	CREATE_ONLY   = "io.rancher.os.createonly"
	RELOAD_CONFIG = "io.rancher.os.reloadconfig"
	CONSOLE       = "io.rancher.os.console"
	SCOPE         = "io.rancher.os.scope"
	REBUILD       = "io.docker.compose.rebuild"
	SYSTEM        = "system"

	OsConfigFile           = "/usr/share/ros/os-config.yml"
	CloudConfigDir         = "/var/lib/rancher/conf/cloud-config.d"
	CloudConfigBootFile    = "/var/lib/rancher/conf/cloud-config.d/boot.yml"
	CloudConfigNetworkFile = "/var/lib/rancher/conf/cloud-config.d/network.yml"
	CloudConfigScriptFile  = "/var/lib/rancher/conf/cloud-config-script"
	MetaDataFile           = "/var/lib/rancher/conf/metadata"
	CloudConfigFile        = "/var/lib/rancher/conf/cloud-config.yml"
)

var (
	OemConfigFile = OEM + "/oem-config.yml"
	VERSION       string
	ARCH          string
	SUFFIX        string
	OS_REPO       string
	OS_BASE       string
	PrivateKeys   = []string{
		"rancher.ssh",
		"rancher.docker.ca_key",
		"rancher.docker.ca_cert",
		"rancher.docker.server_key",
		"rancher.docker.server_cert",
	}
)

func init() {
	if VERSION == "" {
		VERSION = "v0.0.0-dev"
	}
	if ARCH == "" {
		ARCH = runtime.GOARCH
	}
	if SUFFIX == "" && ARCH != "amd64" {
		SUFFIX = "_" + ARCH
	}
	if OS_BASE == "" {
		OS_BASE = fmt.Sprintf("%s/os-base:%s%s", OS_REPO, VERSION, SUFFIX)
	}
}

type Repository struct {
	Url string `yaml:"url,omitempty"`
}

type Repositories map[string]Repository

type CloudConfig struct {
	SSHAuthorizedKeys []string              `yaml:"ssh_authorized_keys"`
	WriteFiles        []File                `yaml:"write_files"`
	Hostname          string                `yaml:"hostname"`
	Mounts            [][]string            `yaml:"mounts,omitempty"`
	Rancher           RancherConfig         `yaml:"rancher,omitempty"`
	Runcmd            []yaml.StringandSlice `yaml:"runcmd,omitempty"`
	Bootcmd           []yaml.StringandSlice `yaml:"bootcmd,omitempty"`
}

type File struct {
	config.File
	Container string `yaml:"container,omitempty"`
}

type RancherConfig struct {
	Console             string                                    `yaml:"console,omitempty"`
	Environment         map[string]string                         `yaml:"environment,omitempty"`
	Services            map[string]*composeConfig.ServiceConfigV1 `yaml:"services,omitempty"`
	BootstrapContainers map[string]*composeConfig.ServiceConfigV1 `yaml:"bootstrap,omitempty"`
	CloudInitServices   map[string]*composeConfig.ServiceConfigV1 `yaml:"cloud_init_services,omitempty"`
	BootstrapDocker     DockerConfig                              `yaml:"bootstrap_docker,omitempty"`
	CloudInit           CloudInit                                 `yaml:"cloud_init,omitempty"`
	Debug               bool                                      `yaml:"debug,omitempty"`
	RmUsr               bool                                      `yaml:"rm_usr,omitempty"`
	NoSharedRoot        bool                                      `yaml:"no_sharedroot,omitempty"`
	Log                 bool                                      `yaml:"log,omitempty"`
	ForceConsoleRebuild bool                                      `yaml:"force_console_rebuild,omitempty"`
	Disable             []string                                  `yaml:"disable,omitempty"`
	ServicesInclude     map[string]bool                           `yaml:"services_include,omitempty"`
	Modules             []string                                  `yaml:"modules,omitempty"`
	Network             NetworkConfig                             `yaml:"network,omitempty"`
	DefaultNetwork      NetworkConfig                             `yaml:"default_network,omitempty"`
	Repositories        Repositories                              `yaml:"repositories,omitempty"`
	Ssh                 SshConfig                                 `yaml:"ssh,omitempty"`
	State               StateConfig                               `yaml:"state,omitempty"`
	SystemDocker        DockerConfig                              `yaml:"system_docker,omitempty"`
	Upgrade             UpgradeConfig                             `yaml:"upgrade,omitempty"`
	Docker              DockerConfig                              `yaml:"docker,omitempty"`
	RegistryAuths       map[string]types.AuthConfig               `yaml:"registry_auths,omitempty"`
	Defaults            Defaults                                  `yaml:"defaults,omitempty"`
	ResizeDevice        string                                    `yaml:"resize_device,omitempty"`
	Sysctl              map[string]string                         `yaml:"sysctl,omitempty"`
	RestartServices     []string                                  `yaml:"restart_services,omitempty"`
}

type UpgradeConfig struct {
	Url      string `yaml:"url,omitempty"`
	Image    string `yaml:"image,omitempty"`
	Rollback string `yaml:"rollback,omitempty"`
}

type EngineOpts struct {
	Bridge           string            `yaml:"bridge,omitempty" opt:"bridge"`
	ConfigFile       string            `yaml:"config_file,omitempty" opt:"config-file"`
	Containerd       string            `yaml:"containerd,omitempty" opt:"containerd"`
	Debug            *bool             `yaml:"debug,omitempty" opt:"debug"`
	ExecRoot         string            `yaml:"exec_root,omitempty" opt:"exec-root"`
	Group            string            `yaml:"group,omitempty" opt:"group"`
	Graph            string            `yaml:"graph,omitempty" opt:"graph"`
	Host             []string          `yaml:"host,omitempty" opt:"host"`
	InsecureRegistry []string          `yaml:"insecure_registry" opt:"insecure-registry"`
	LiveRestore      *bool             `yaml:"live_restore,omitempty" opt:"live-restore"`
	LogDriver        string            `yaml:"log_driver,omitempty" opt:"log-driver"`
	LogOpts          map[string]string `yaml:"log_opts,omitempty" opt:"log-opt"`
	PidFile          string            `yaml:"pid_file,omitempty" opt:"pidfile"`
	RegistryMirror   string            `yaml:"registry_mirror,omitempty" opt:"registry-mirror"`
	Restart          *bool             `yaml:"restart,omitempty" opt:"restart"`
	SelinuxEnabled   *bool             `yaml:"selinux_enabled,omitempty" opt:"selinux-enabled"`
	StorageDriver    string            `yaml:"storage_driver,omitempty" opt:"storage-driver"`
	UserlandProxy    *bool             `yaml:"userland_proxy,omitempty" opt:"userland-proxy"`
}

type DockerConfig struct {
	EngineOpts
	Engine         string   `yaml:"engine,omitempty"`
	TLS            bool     `yaml:"tls,omitempty"`
	TLSArgs        []string `yaml:"tls_args,flow,omitempty"`
	ExtraArgs      []string `yaml:"extra_args,flow,omitempty"`
	ServerCert     string   `yaml:"server_cert,omitempty"`
	ServerKey      string   `yaml:"server_key,omitempty"`
	CACert         string   `yaml:"ca_cert,omitempty"`
	CAKey          string   `yaml:"ca_key,omitempty"`
	Environment    []string `yaml:"environment,omitempty"`
	StorageContext string   `yaml:"storage_context,omitempty"`
	Exec           bool     `yaml:"exec,omitempty"`
}

type NetworkConfig struct {
	PreCmds    []string                   `yaml:"pre_cmds,omitempty"`
	Dns        DnsConfig                  `yaml:"dns,omitempty"`
	Interfaces map[string]InterfaceConfig `yaml:"interfaces,omitempty"`
	PostCmds   []string                   `yaml:"post_cmds,omitempty"`
	HttpProxy  string                     `yaml:"http_proxy,omitempty"`
	HttpsProxy string                     `yaml:"https_proxy,omitempty"`
	NoProxy    string                     `yaml:"no_proxy,omitempty"`
}

type InterfaceConfig struct {
	Match       string            `yaml:"match,omitempty"`
	DHCP        bool              `yaml:"dhcp,omitempty"`
	DHCPArgs    string            `yaml:"dhcp_args,omitempty"`
	Address     string            `yaml:"address,omitempty"`
	Addresses   []string          `yaml:"addresses,omitempty"`
	IPV4LL      bool              `yaml:"ipv4ll,omitempty"`
	Gateway     string            `yaml:"gateway,omitempty"`
	GatewayIpv6 string            `yaml:"gateway_ipv6,omitempty"`
	MTU         int               `yaml:"mtu,omitempty"`
	Bridge      string            `yaml:"bridge,omitempty"`
	Bond        string            `yaml:"bond,omitempty"`
	BondOpts    map[string]string `yaml:"bond_opts,omitempty"`
	PostUp      []string          `yaml:"post_up,omitempty"`
	PreUp       []string          `yaml:"pre_up,omitempty"`
	Vlans       string            `yaml:"vlans,omitempty"`
}

type DnsConfig struct {
	Nameservers []string `yaml:"nameservers,flow,omitempty"`
	Search      []string `yaml:"search,flow,omitempty"`
}

type SshConfig struct {
	Keys map[string]string `yaml:"keys,omitempty"`
}

type StateConfig struct {
	Directory  string   `yaml:"directory,omitempty"`
	FsType     string   `yaml:"fstype,omitempty"`
	Dev        string   `yaml:"dev,omitempty"`
	Wait       bool     `yaml:"wait,omitempty"`
	Required   bool     `yaml:"required,omitempty"`
	Autoformat []string `yaml:"autoformat,omitempty"`
	MdadmScan  bool     `yaml:"mdadm_scan,omitempty"`
	Script     string   `yaml:"script,omitempty"`
	OemFsType  string   `yaml:"oem_fstype,omitempty"`
	OemDev     string   `yaml:"oem_dev,omitempty"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}

type Defaults struct {
	Hostname string        `yaml:"hostname,omitempty"`
	Docker   DockerConfig  `yaml:"docker,omitempty"`
	Network  NetworkConfig `yaml:"network,omitempty"`
}

func (r Repositories) ToArray() []string {
	result := make([]string, 0, len(r))
	for _, repo := range r {
		if repo.Url != "" {
			result = append(result, repo.Url)
		}
	}

	return result
}
