package config

import (
	"fmt"
	"runtime"

	"github.com/rancher/os/config/cloudinit/config"
	"github.com/rancher/os/config/yaml"
	"github.com/rancher/os/pkg/netconf"

	"github.com/docker/engine-api/types"
	composeConfig "github.com/docker/libcompose/config"
)

const (
	OemDir           = "/usr/share/ros/oem"
	BootDir          = "/boot"
	StateDir         = "/state"
	RosBin           = "/usr/bin/ros"
	SysInitBin       = "/usr/bin/ros-sysinit"
	SystemDockerHost = "unix:///var/run/system-docker.sock"
	DockerHost       = "unix:///var/run/docker.sock"
	ImagesPath       = "/usr/share/ros"
	InitImages       = "images-init.tar"
	SystemImages     = "images-system.tar"
	Debug            = false
	SystemDockerBin  = "/usr/bin/system-dockerd"
	DefaultDind      = "rancher/os-dind:17.12.1"

	DetachLabel        = "io.rancher.os.detach"
	CreateOnlyLabel    = "io.rancher.os.createonly"
	ReloadConfigLabel  = "io.rancher.os.reloadconfig"
	ConsoleLabel       = "io.rancher.os.console"
	ScopeLabel         = "io.rancher.os.scope"
	RebuildLabel       = "io.docker.compose.rebuild"
	UserDockerLabel    = "io.rancher.user_docker.name"
	UserDockerNetLabel = "io.rancher.user_docker.net"
	UserDockerFIPLabel = "io.rancher.user_docker.fix_ip"
	System             = "system"

	OsConfigFile           = "/usr/share/ros/os-config.yml"
	VarRancherDir          = "/var/lib/rancher"
	CloudConfigDir         = "/var/lib/rancher/conf/cloud-config.d"
	CloudConfigInitFile    = "/var/lib/rancher/conf/cloud-config.d/init.yml"
	CloudConfigBootFile    = "/var/lib/rancher/conf/cloud-config.d/boot.yml"
	CloudConfigNetworkFile = "/var/lib/rancher/conf/cloud-config.d/network.yml"
	CloudConfigScriptFile  = "/var/lib/rancher/conf/cloud-config-script"
	MetaDataFile           = "/var/lib/rancher/conf/metadata"
	CloudConfigFile        = "/var/lib/rancher/conf/cloud-config.yml"
	EtcResolvConfFile      = "/etc/resolv.conf"
	WPAConfigFile          = "/etc/wpa_supplicant-%s.conf"
	WPATemplateFile        = "/etc/wpa_supplicant.conf.tpl"
	DHCPCDConfigFile       = "/etc/dhcpcd.conf"
	DHCPCDTemplateFile     = "/etc/dhcpcd.conf.tpl"
	MultiDockerConfFile    = "/var/lib/rancher/conf.d/m-user-docker.yml"
	MultiDockerDataDir     = "/var/lib/m-user-docker"
	UdevRulesDir           = "/etc/udev/rules.d"
	UdevRulesExtrasDir     = "/lib/udev/rules-extras.d"
)

var (
	OemConfigFile = OemDir + "/oem-config.yml"
	Version       string
	BuildDate     string
	Arch          string
	Suffix        string
	OsRepo        string
	OsBase        string
	PrivateKeys   = []string{
		"rancher.ssh",
		"rancher.docker.ca_key",
		"rancher.docker.ca_cert",
		"rancher.docker.server_key",
		"rancher.docker.server_cert",
	}
	Additional = []string{
		"rancher.password",
		"rancher.autologin",
		"EXTRA_CMDLINE",
	}
	SupportedDinds = []string{
		"rancher/os-dind:17.12.1",
		"rancher/os-dind:18.03.1",
	}
)

func init() {
	if Version == "" {
		Version = "v0.0.0-dev"
	}
	if Arch == "" {
		Arch = runtime.GOARCH
	}
	if Suffix == "" && Arch != "amd64" {
		Suffix = "_" + Arch
	}
	if OsBase == "" {
		OsBase = fmt.Sprintf("%s/os-base:%s%s", OsRepo, Version, Suffix)
	}
}

type Repository struct {
	URL string `yaml:"url,omitempty"`
}

type Repositories map[string]Repository

type CloudConfig struct {
	SSHAuthorizedKeys []string              `yaml:"ssh_authorized_keys,omitempty"`
	WriteFiles        []File                `yaml:"write_files,omitempty"`
	Hostname          string                `yaml:"hostname,omitempty"`
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
	Recovery            bool                                      `yaml:"recovery,omitempty"`
	Disable             []string                                  `yaml:"disable,omitempty"`
	ServicesInclude     map[string]bool                           `yaml:"services_include,omitempty"`
	Modules             []string                                  `yaml:"modules,omitempty"`
	Network             netconf.NetworkConfig                     `yaml:"network,omitempty"`
	Repositories        Repositories                              `yaml:"repositories,omitempty"`
	SSH                 SSHConfig                                 `yaml:"ssh,omitempty"`
	State               StateConfig                               `yaml:"state,omitempty"`
	SystemDocker        DockerConfig                              `yaml:"system_docker,omitempty"`
	Upgrade             UpgradeConfig                             `yaml:"upgrade,omitempty"`
	Docker              DockerConfig                              `yaml:"docker,omitempty"`
	RegistryAuths       map[string]types.AuthConfig               `yaml:"registry_auths,omitempty"`
	Defaults            Defaults                                  `yaml:"defaults,omitempty"`
	ResizeDevice        string                                    `yaml:"resize_device,omitempty"`
	Sysctl              map[string]string                         `yaml:"sysctl,omitempty"`
	RestartServices     []string                                  `yaml:"restart_services,omitempty"`
	HypervisorService   bool                                      `yaml:"hypervisor_service,omitempty"`
	ShutdownTimeout     int                                       `yaml:"shutdown_timeout,omitempty"`
	PreloadWait         bool                                      `yaml:"preload_wait,omitempty"`
}

type UpgradeConfig struct {
	URL      string `yaml:"url,omitempty"`
	Image    string `yaml:"image,omitempty"`
	Rollback string `yaml:"rollback,omitempty"`
	Policy   string `yaml:"policy,omitempty"`
}

type EngineOpts struct {
	Bridge           string            `yaml:"bridge,omitempty" opt:"bridge"`
	BIP              string            `yaml:"bip,omitempty" opt:"bip"`
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

type SSHConfig struct {
	Keys          map[string]string `yaml:"keys,omitempty"`
	Daemon        bool              `yaml:"daemon,omitempty"`
	Port          int               `yaml:"port,omitempty"`
	ListenAddress string            `yaml:"listen_address,omitempty"`
}

type StateConfig struct {
	Directory  string   `yaml:"directory,omitempty"`
	FsType     string   `yaml:"fstype,omitempty"`
	Dev        string   `yaml:"dev,omitempty"`
	Wait       bool     `yaml:"wait,omitempty"`
	Required   bool     `yaml:"required,omitempty"`
	Autoformat []string `yaml:"autoformat,omitempty"`
	MdadmScan  bool     `yaml:"mdadm_scan,omitempty"`
	LvmScan    bool     `yaml:"lvm_scan,omitempty"`
	Cryptsetup bool     `yaml:"cryptsetup,omitempty"`
	Rngd       bool     `yaml:"rngd,omitempty"`
	Script     string   `yaml:"script,omitempty"`
	OemFsType  string   `yaml:"oem_fstype,omitempty"`
	OemDev     string   `yaml:"oem_dev,omitempty"`
	BootFsType string   `yaml:"boot_fstype,omitempty"`
	BootDev    string   `yaml:"boot_dev,omitempty"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}

type Defaults struct {
	Hostname         string                `yaml:"hostname,omitempty"`
	Docker           DockerConfig          `yaml:"docker,omitempty"`
	Network          netconf.NetworkConfig `yaml:"network,omitempty"`
	SystemDockerLogs string                `yaml:"system_docker_logs,omitempty"`
}

func (r Repositories) ToArray() []string {
	result := make([]string, 0, len(r))
	for _, repo := range r {
		if repo.URL != "" {
			result = append(result, repo.URL)
		}
	}

	return result
}
