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

type InitFunc func(*Config) error

type ContainerConfig struct {
	Id  string `yaml:"id,omitempty"`
	Cmd string `yaml:"run,omitempty"`
	//Config     *runconfig.Config     `yaml:"-"`
	//HostConfig *runconfig.HostConfig `yaml:"-"`
}

type Config struct {
	//BootstrapContainers []ContainerConfig `yaml:"bootstrapContainers,omitempty"`
	//UserContainers   []ContainerConfig `yaml:"userContainser,omitempty"`
	Debug            bool              `yaml:"debug,omitempty"`
	Disable          []string          `yaml:"disable,omitempty"`
	Dns              []string          `yaml:"dns,omitempty"`
	Rescue           bool              `yaml:"rescue,omitempty"`
	RescueContainer  *ContainerConfig  `yaml:"rescue_container,omitempty"`
	State            ConfigState       `yaml:"state,omitempty"`
	SystemContainers []ContainerConfig `yaml:"system_containers,omitempty"`
	SystemDockerArgs []string          `yaml:"system_docker_args,flow,omitempty"`
	Modules          []string          `yaml:"modules,omitempty"`
}

type ConfigState struct {
	FsType   string `yaml:"fstype"`
	Dev      string `yaml:"dev"`
	Required bool   `yaml:"required"`
}

func (c *Config) ClearReadOnly() {
	c.SystemContainers = []ContainerConfig{}
	c.RescueContainer = nil
}

func (c *Config) Dump() (string, error) {
	content, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	} else {
		return string(content), err
	}
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

func (c *Config) readArgs() error {
	log.Debug("Reading config args")
	cmdLine := strings.Join(os.Args[1:], " ")
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
	return yaml.Unmarshal(override, c)
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

func dummyMarshall(value string) interface{} {
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
				if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
					current[key] = strings.Split(value[1:len(value)-1], ",")
				} else {
					current[key] = dummyMarshall(value)
				}
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
		c.readCmdline,
		c.readArgs,
	)
}

func (c *Config) GetContainerById(id string) *ContainerConfig {
	for _, c := range c.SystemContainers {
		if c.Id == id {
			return &c
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
