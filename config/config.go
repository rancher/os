package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/util"
)

type InitFunc func(*Config) error

type ContainerConfig struct {
	Id  string   `json:"id,omitempty"`
	Cmd []string `json:"run,omitempty"`
	//Config     *runconfig.Config     `json:"-"`
	//HostConfig *runconfig.HostConfig `json:"-"`
}

type Config struct {
	//BootstrapContainers []ContainerConfig `json:"bootstrapContainers,omitempty"`
	ConsoleContainer string            `json:"consoleContainer,omitempty"`
	Debug            bool              `json:"debug,omitempty"`
	Disable          []string          `json:"disable,omitempty"`
	DockerEndpoint   string            `json:"dockerEndpoint,omitempty"`
	Dns              []string          `json:"dns,omitempty"`
	ImagesPath       string            `json:"ImagesPath,omitempty"`
	ImagesPattern    string            `json:"ImagesPattern,omitempty"`
	ModulesArchive   string            `json:"modulesArchive,omitempty"`
	Rescue           bool              `json:"rescue,omitempty"`
	RescueContainer  ContainerConfig   `json:"rescueContainer,omitempty"`
	StateDevFSType   string            `json:"stateDeviceFsType,omitempty"`
	StateDev         string            `json:"stateDevice,omitempty"`
	StateRequired    bool              `json:"stateRequired,omitempty"`
	SysInit          string            `json:"sysInit,omitempty"`
	SystemContainers []ContainerConfig `json:"systemContainers,omitempty"`
	SystemDockerArgs []string          `json:"systemDockerArgs,omitempty"`
	UserContainers   []ContainerConfig `json:"userContainser,omitempty"`
	UserInit         string            `json:"userInit,omitempty"`
	DockerBin        string            `json:"dockerBin,omitempty"`
	Modules          []string          `json:"modules,omitempty"`
	Respawn          []string          `json:"respawn,omitempty"`
}

func (c *Config) Dump() string {
	content, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		return string(content)
	} else {
		return err.Error()
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

func NewConfig() *Config {
	return &Config{
		ConsoleContainer: "console",
		DockerBin:        "/usr/bin/docker",
		Debug:            true,
		DockerEndpoint:   "unix:/var/run/docker.sock",
		Dns: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
		ImagesPath:    "/",
		ImagesPattern: "images*.tar",
		StateRequired: false,
		StateDev:         "LABEL=RANCHER_STATE",
		StateDevFSType:   "auto",
		SysInit:          "/sbin/init-sys",
		SystemDockerArgs: []string{"docker", "-d", "-s", "overlay", "-b", "none"},
		UserInit:         "/sbin/init-user",
		Modules:          []string{},
		ModulesArchive:   "/modules.tar",
		SystemContainers: []ContainerConfig{
			{
				Cmd: []string{
					"--name", "system-state",
					"--net", "none",
					"--read-only",
					"state",
				},
			},
			{
				Cmd: []string{
					"--name", "udev",
					"--net", "none",
					"--privileged",
					"--rm",
					"--volume", "/dev:/host/dev",
					"--volume", "/lib/modules:/lib/modules:ro",
					"udev",
				},
			},
			{
				Cmd: []string{
					"--name", "network",
					"--cap-add", "NET_ADMIN",
					"--net", "host",
					"--rm",
					"network",
				},
			},
			{
				Cmd: []string{
					"--name", "userdocker",
					"-d",
					"--restart", "always",
					"--pid", "host",
					"--net", "host",
					"--privileged",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volumes-from", "system-state",
					"userdocker",
				},
			},
			{
				Cmd: []string{
					"--name", "console",
					"-d",
					"--rm",
					"--privileged",
					//"--volume", "/:/host:ro",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volume", "/init:/usr/bin/system-docker:ro",
					"--volume", "/init:/usr/bin/respawn:ro",
					"--volume", "/var/run/docker.sock:/var/run/system-docker.sock:ro",
					"--volumes-from", "system-state",
					"--net", "host",
					"--pid", "host",
					"console",
				},
			},
		},
		RescueContainer: ContainerConfig{
			Cmd: []string{
				"--name", "rescue",
				"-d",
				"--rm",
				"--privileged",
				//"--volume", "/:/host",
				"--volume", "/lib/modules:/lib/modules:ro",
				"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
				"--volume", "/init:/usr/bin/system-docker:ro",
				"--volume", "/init:/usr/bin/respawn:ro",
				"--volume", "/var/run/docker.sock:/var/run/system-docker.sock:ro",
				"--net", "host",
				"--pid", "host",
				"rescue",
			},
		},
	}
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
	override, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return json.Unmarshal(override, c)
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
