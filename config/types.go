package config

import (
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/docker/libcompose/project"
	"github.com/rancher/netconf"
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

	LABEL         = "label"
	HASH          = "io.rancher.os.hash"
	ID            = "io.rancher.os.id"
	DETACH        = "io.rancher.os.detach"
	CREATE_ONLY   = "io.rancher.os.createonly"
	RELOAD_CONFIG = "io.rancher.os.reloadconfig"
	SCOPE         = "io.rancher.os.scope"
	REBUILD       = "io.docker.compose.rebuild"
	SYSTEM        = "system"

	OsConfigFile           = "/usr/share/ros/os-config.yml"
	CloudConfigDir         = "/var/lib/rancher/conf/cloud-config.d"
	CloudConfigBootFile    = "/var/lib/rancher/conf/cloud-config.d/boot.yml"
	CloudConfigPrivateFile = "/var/lib/rancher/conf/cloud-config.d/private.yml"
	CloudConfigNetworkFile = "/var/lib/rancher/conf/cloud-config.d/network.yml"
	CloudConfigScriptFile  = "/var/lib/rancher/conf/cloud-config-script"
	MetaDataFile           = "/var/lib/rancher/conf/metadata"
	CloudConfigFile        = "/var/lib/rancher/conf/cloud-config.yml"
)

var (
	OemConfigFile = OEM + "/oem-config.yml"
	VERSION       string
)

func init() {
	if VERSION == "" {
		VERSION = "v0.0.0-dev"
	}
}

type Repository struct {
	Url string `yaml:"url,omitempty"`
}

type Repositories map[string]Repository

type CloudConfig struct {
	SSHAuthorizedKeys []string      `yaml:"ssh_authorized_keys"`
	WriteFiles        []config.File `yaml:"write_files"`
	Hostname          string        `yaml:"hostname"`

	Rancher RancherConfig `yaml:"rancher,omitempty"`
}

type RancherConfig struct {
	Environment         map[string]string                 `yaml:"environment,omitempty"`
	Services            map[string]*project.ServiceConfig `yaml:"services,omitempty"`
	BootstrapContainers map[string]*project.ServiceConfig `yaml:"bootstrap,omitempty"`
	Autoformat          map[string]*project.ServiceConfig `yaml:"autoformat,omitempty"`
	BootstrapDocker     DockerConfig                      `yaml:"bootstrap_docker,omitempty"`
	CloudInit           CloudInit                         `yaml:"cloud_init,omitempty"`
	Debug               bool                              `yaml:"debug,omitempty"`
	RmUsr               bool                              `yaml:"rm_usr,omitempty"`
	Log                 bool                              `yaml:"log,omitempty"`
	Disable             []string                          `yaml:"disable,omitempty"`
	ServicesInclude     map[string]bool                   `yaml:"services_include,omitempty"`
	Modules             []string                          `yaml:"modules,omitempty"`
	Network             netconf.NetworkConfig             `yaml:"network,omitempty"`
	Repositories        Repositories                      `yaml:"repositories,omitempty"`
	State               StateConfig                       `yaml:"state,omitempty"`
	SystemDocker        DockerConfig                      `yaml:"system_docker,omitempty"`
	Upgrade             UpgradeConfig                     `yaml:"upgrade,omitempty"`
	Docker              DockerConfig                      `yaml:"docker,omitempty"`
}

type UpgradeConfig struct {
	Url      string `yaml:"url,omitempty"`
	Image    string `yaml:"image,omitempty"`
	Rollback string `yaml:"rollback,omitempty"`
}

type DockerConfig struct {
	TLS            bool     `yaml:"tls,omitempty"`
	TLSArgs        []string `yaml:"tls_args,flow,omitempty"`
	Args           []string `yaml:"args,flow,omitempty"`
	ExtraArgs      []string `yaml:"extra_args,flow,omitempty"`
	ServerCert     string   `yaml:"server_cert,omitempty"`
	ServerKey      string   `yaml:"server_key,omitempty"`
	CACert         string   `yaml:"ca_cert,omitempty"`
	CAKey          string   `yaml:"ca_key,omitempty"`
	Environment    []string `yaml:"environment,omitempty"`
	StorageContext string   `yaml:"storage_context,omitempty"`
	Exec           bool     `yaml:"exec,omitempty"`
}

type StateConfig struct {
	Directory  string   `yaml:"directory,omitempty"`
	FsType     string   `yaml:"fstype,omitempty"`
	Dev        string   `yaml:"dev,omitempty"`
	Required   bool     `yaml:"required,omitempty"`
	Autoformat []string `yaml:"autoformat,omitempty"`
	FormatZero bool     `yaml:"formatzero,omitempty"`
	MdadmScan  bool     `yaml:"mdadm_scan,omitempty"`
	Script     string   `yaml:"script,omitempty"`
	OemFsType  string   `yaml:"oem_fstype,omitempty"`
	OemDev     string   `yaml:"oem_dev,omitempty"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
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
