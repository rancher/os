package config

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type InitFunc func(*Config) error

type ContainerConfig struct {
	Options []string `json:"options,omitempty"`
	Image   string   `json:"image,omitempty"`
	Args    []string `json:"args,omitempty"`
}

type Config struct {
	BootstrapContainers []ContainerConfig `json:"bootstrapContainers,omitempty"`
	Debug               bool              `json:"debug,omitempty"`
	DockerEndpoint      string            `json:"dockerEndpoint,omitempty"`
	Dns                 []string          `json:"dns,omitempty"`
	ImagesPath          string            `json:"ImagesPath,omitempty"`
	ImagesPattern       string            `json:"ImagesPattern,omitempty"`
	ModulesArchive      string            `json:"modulesArchive,omitempty"`
	Rescue              bool              `json:"rescue,omitempty"`
	RescueContainer     ContainerConfig   `json:"rescueContainer,omitempty"`
	StateDevFSType      string            `json:"stateDeviceFsType,omitempty"`
	StateDev            string            `json:"stateDevice,omitempty"`
	StateRequired       bool              `json:"stateRequired,omitempty"`
	SysInit             string            `json:"sysInit,omitempty"`
	SystemContainers    []ContainerConfig `json:"systemContainers,omitempty"`
	SystemDockerArgs    []string          `json:"systemDockerArgs,omitempty"`
	UserContainers      []ContainerConfig `json:"userContainser,omitempty"`
	UserInit            string            `json:"userInit,omitempty"`
	DockerBin           string            `json:"dockerBin,omitempty"`
	Modules             []string          `json:"modules,omitempty"`
	Respawn             []string          `json:"respawn,omitempty"`
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
		DockerBin:      "/usr/bin/docker",
		Debug:          true,
		DockerEndpoint: "unix:/var/run/docker.sock",
		Dns: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
		ImagesPath:    "/",
		ImagesPattern: "images*.tar",
		StateRequired: false,
		//StateDev:         "/dev/sda",
		StateDevFSType:   "ext4",
		SysInit:          "/sbin/init-sys",
		SystemDockerArgs: []string{"docker", "-d", "-s", "overlay", "-b", "none"},
		UserInit:         "/sbin/init-user",
		Modules:          []string{},
		ModulesArchive:   "/modules.tar",
		SystemContainers: []ContainerConfig{
			{
				Options: []string{
					"--name", "system-state",
					"--net", "none",
					"--read-only",
				},
				Image: "state",
			},
			{
				Options: []string{
					"--net", "none",
					"--privileged",
					"--rm",
					"--volume", "/dev:/host/dev",
					"--volume", "/lib/modules:/lib/modules:ro",
				},
				Image: "udev",
			},
			{
				Options: []string{
					"--cap-add", "NET_ADMIN",
					"--net", "host",
					"--rm",
				},
				Image: "network",
			},
			{
				Options: []string{
					"-d",
					"--restart", "always",
					"--net", "host",
					"--privileged",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volumes-from", "system-state",
				},
				Image: "userdocker",
			},
			{
				Options: []string{
					"--rm",
					"--privileged",
					"--volume", "/:/host:ro",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volume", "/usr/bin/system-docker:/usr/bin/system-docker:ro",
					"--volume", "/var/run/docker.sock:/var/run/system-docker.sock:ro",
					"--volumes-from", "system-state",
					"--net", "host",
					"--pid", "host",
					"-it",
				},
				Image: "console",
			},
		},
		RescueContainer: ContainerConfig{
			Options: []string{
				"--rm",
				"--privileged",
				"--volume", "/:/host",
				"--volume", "/lib/modules:/lib/modules:ro",
				"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
				"--volume", "/var/run/docker.sock:/var/run/docker.sock:ro",
				"--net", "host",
				"--pid", "host",
				"-it",
			},
			Image: "rescue",
		},
	}
}

func (c *Config) readCmdline() error {
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))

	// Lazy way to assign values to *Config
	b, err := json.Marshal(cmdLineObj)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c)
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
				current[key] = dummyMarshall(value)
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

	return result
}

func (c *Config) Reload() error {
	return c.readCmdline()
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
