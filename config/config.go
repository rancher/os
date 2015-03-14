package config

import (
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/util"
	"gopkg.in/yaml.v2"
)

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
	ConfigFile = "/var/lib/rancher/conf/rancher.yml"
)

type InitFunc func(*Config) error

type ContainerConfig struct {
	Id             string `yaml:"id,omitempty"`
	Cmd            string `yaml:"run,omitempty"`
	MigrateVolumes bool   `yaml:"migrate_volumes,omitempty"`
	ReloadConfig   bool   `yaml:"reload_config,omitempty"`
}

type Config struct {
	Debug            bool              `yaml:"debug,omitempty"`
	Disable          []string          `yaml:"disable,omitempty"`
	Dns              []string          `yaml:"dns,flow,omitempty"`
	Rescue           bool              `yaml:"rescue,omitempty"`
	RescueContainer  *ContainerConfig  `yaml:"rescue_container,omitempty"`
	State            ConfigState       `yaml:"state,omitempty"`
	Userdocker       UserDockerInfo    `yaml:"userdocker,omitempty"`
	OsUpgradeChannel string            `yaml:"os_upgrade_channel,omitempty"`
	SystemContainers []ContainerConfig `yaml:"system_containers,omitempty"`
	SystemDockerArgs []string          `yaml:"system_docker_args,flow,omitempty"`
	Modules          []string          `yaml:"modules,omitempty"`
	CloudInit        CloudInit         `yaml:"cloud_init"`
	SshInfo          SshInfo           `yaml:"ssh"`
	EnabledAddons    []string          `yaml:"enabledAddons,omitempty"`
	Addons           map[string]Config `yaml:"addons,omitempty"`
	Network          NetworkConfig     `yaml:"network,omitempty"`
}

type NetworkConfig struct {
	Interfaces []InterfaceConfig `yaml:"interfaces"`
	PostRun    *ContainerConfig  `yaml:"post_run"`
}

type InterfaceConfig struct {
	Match   string `yaml:"match"`
	DHCP    bool   `yaml:"dhcp"`
	Address string `yaml:"address,omitempty"`
	Gateway string `yaml:"gateway,omitempty"`
	MTU     int    `yaml:"mtu,omitempty"`
}

type UserDockerInfo struct {
	UseTLS        bool   `yaml:"use_tls"`
	TLSServerCert string `yaml:"tls_server_cert"`
	TLSServerKey  string `yaml:"tls_server_key"`
	TLSCACert     string `yaml:"tls_ca_cert"`
}

type SshInfo struct {
	Keys map[string]string
}

type ConfigState struct {
	FsType   string `yaml:"fstype"`
	Dev      string `yaml:"dev"`
	Required bool   `yaml:"required"`
}

type CloudInit struct {
	Datasources []string `yaml:"datasources,omitempty"`
}

func (c *Config) PrivilegedMerge(newConfig Config) (bool, error) {
	reboot, err := c.Merge(newConfig)
	if err != nil {
		return reboot, err
	}

	toAppend := make([]ContainerConfig, 0, 5)

	for _, newContainer := range newConfig.SystemContainers {
		found := false
		for i, existingContainer := range c.SystemContainers {
			if existingContainer.Id != "" && newContainer.Id == existingContainer.Id {
				found = true
				c.SystemContainers[i] = newContainer
			}
		}
		if !found {
			toAppend = append(toAppend, newContainer)
		}
	}

	c.SystemContainers = append(c.SystemContainers, toAppend...)

	return reboot, nil
}

func (c *Config) Merge(newConfig Config) (bool, error) {
	//Efficient? Nope, but computers are fast
	newConfig.ClearReadOnly()
	content, err := newConfig.Dump()
	if err != nil {
		return false, err
	}

	log.Debugf("Input \n%s", string(content))

	err = yaml.Unmarshal([]byte(content), c)
	return true, err
}

func (c *Config) ClearReadOnly() {
	c.SystemContainers = []ContainerConfig{}
	c.RescueContainer = nil
	c.Rescue = false
	c.Debug = false
	c.Addons = map[string]Config{}
}

func (c *Config) Dump() (string, error) {
	content, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	} else {
		return string(content), err
	}
}

func (c Config) Save() error {
	c.ClearReadOnly()
	content, err := c.Dump()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(ConfigFile, []byte(content), 400)
}

func LoadConfig() (*Config, error) {
	cfg := NewConfig()
	if err := cfg.Reload(); err != nil {
		return nil, err
	}

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	return cfg, nil
}

func FilterGlobalConfig(input []string) []string {
	result := make([]string, 0, len(input))
	for _, value := range input {
		if !strings.HasPrefix(value, "--rancher") {
			result = append(result, value)
		}
	}

	return result
}

func (c *Config) readArgs() error {
	log.Debug("Reading config args")
	parts := make([]string, len(os.Args))

	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") {
			arg = arg[2:]
		}

		kv := strings.SplitN(arg, "=", 2)
		kv[0] = strings.Replace(kv[0], "-", ".", -1)
		parts = append(parts, strings.Join(kv, "="))
	}

	cmdLine := strings.Join(parts, " ")
	if len(cmdLine) == 0 {
		return nil
	}

	log.Debugf("Config Args %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(cmdLine))

	return c.merge(cmdLineObj)
}

func (c *Config) merge(values map[string]interface{}) error {
	// Lazy way to assign values to *Config
	override, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	var newConfig Config
	err = yaml.Unmarshal(override, &newConfig)
	if err != nil {
		return nil
	}

	_, err = c.Merge(newConfig)
	return err
}

func (c *Config) readFile() error {
	content, err := ioutil.ReadFile(ConfigFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = yaml.Unmarshal(content, data)
	if err != nil {
		return err
	}

	return c.merge(data)
}

func (c *Config) readCmdline() error {
	log.Debug("Reading config cmdline")
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}

	if len(cmdLine) == 0 {
		return nil
	}

	log.Debugf("Config cmdline %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))
	return c.merge(cmdLineObj)
}

func DummyMarshall(value string) interface{} {
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		return strings.Split(value[1:len(value)-1], ",")
	}

	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else if ok, _ := regexp.MatchString("^[0-9]+$", value); ok {
		i, err := strconv.Atoi(value)
		if err != nil {
			panic(err)
		}
		return i
	}

	return value
}

func parseCmdline(cmdLine string) map[string]interface{} {
	result := make(map[string]interface{})

outer:
	for _, part := range strings.Split(cmdLine, " ") {
		if !strings.HasPrefix(part, "rancher.") {
			continue
		}

		var value string
		kv := strings.SplitN(part, "=", 2)

		if len(kv) == 1 {
			value = "true"
		} else {
			value = kv[1]
		}

		current := result
		keys := strings.Split(kv[0], ".")[1:]
		for i, key := range keys {
			if i == len(keys)-1 {
				current[key] = DummyMarshall(value)
			} else {
				if obj, ok := current[key]; ok {
					if newCurrent, ok := obj.(map[string]interface{}); ok {
						current = newCurrent
					} else {
						continue outer
					}
				} else {
					newCurrent := make(map[string]interface{})
					current[key] = newCurrent
					current = newCurrent
				}
			}
		}
	}

	log.Debugf("Input obj %s", result)
	return result
}

func (c *Config) Reload() error {
	return util.ShortCircuit(
		c.readFile,
		c.readCmdline,
		c.readArgs,
		c.mergeAddons,
	)
}

func (c *Config) mergeAddons() error {
	for _, addon := range c.EnabledAddons {
		if newConfig, ok := c.Addons[addon]; ok {
			log.Debugf("Enabling addon %s", addon)
			_, err := c.PrivilegedMerge(newConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RunInitFuncs(cfg *Config, initFuncs []InitFunc) error {
	for i, initFunc := range initFuncs {
		log.Debugf("[%d/%d] Starting", i+1, len(initFuncs))
		if err := initFunc(cfg); err != nil {
			log.Errorf("Failed [%d/%d] %d%%", i+1, len(initFuncs), ((i + 1) * 100 / len(initFuncs)))
			return err
		}
		log.Debugf("[%d/%d] Done %d%%", i+1, len(initFuncs), ((i + 1) * 100 / len(initFuncs)))
	}
	return nil
}
