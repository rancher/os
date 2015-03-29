package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/rancherio/rancher-compose/project"

	log "github.com/Sirupsen/logrus"
	"github.com/rancherio/os/util"
	"gopkg.in/yaml.v2"
)

func (c *Config) privilegedMerge(newConfig Config) error {
	err := c.overlay(newConfig)
	if err != nil {
		return err
	}

	for k, v := range newConfig.SystemContainers {
		c.SystemContainers[k] = v
	}

	return nil
}

func (c *Config) overlay(newConfig Config) error {
	newConfig.clearReadOnly()
	return util.Convert(&newConfig, c)
}

func (c *Config) clearReadOnly() {
	c.BootstrapContainers = make(map[string]*project.ServiceConfig, 0)
	c.SystemContainers = make(map[string]*project.ServiceConfig, 0)
}

func clearReadOnly(data map[interface{}]interface{}) map[interface{}]interface{} {
	newData := make(map[interface{}]interface{})
	for k, v := range data {
		newData[k] = v
	}

	delete(newData, "system_container")
	delete(newData, "bootstrap_container")

	return newData
}

func (c *Config) Import(bytes []byte) error {
	data, err := readConfig(bytes, PrivateConfigFile)
	if err != nil {
		return err
	}

	if err = saveToDisk(data); err != nil {
		return err
	}

	return c.Reload()
}

// This function only sets "non-empty" values
func (c *Config) SetConfig(newConfig *Config) error {
	bytes, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}

	return c.Merge(bytes)
}

func (c *Config) Merge(bytes []byte) error {
	data, err := readSavedConfig(bytes)
	if err != nil {
		return err
	}

	err = saveToDisk(data)
	if err != nil {
		return err
	}

	return c.Reload()
}

func LoadConfig() (*Config, error) {
	cfg := NewConfig()
	if err := cfg.Reload(); err != nil {
		return nil, err
	}

	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		if !util.Contains(cfg.UserDocker.Args, "-D") {
			cfg.UserDocker.Args = append(cfg.UserDocker.Args, "-D")
		}
		if !util.Contains(cfg.SystemDocker.Args, "-D") {
			cfg.SystemDocker.Args = append(cfg.SystemDocker.Args, "-D")
		}
	}

	return cfg, nil
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

func (c *Config) merge(values map[interface{}]interface{}) error {
	values = clearReadOnly(values)
	return util.Convert(values, c)
}

func (c *Config) readFiles() error {
	data, err := readSavedConfig(nil)
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

func Dump(private, full bool) (string, error) {
	files := []string{CloudConfigFile, ConfigFile}
	if private {
		files = append(files, PrivateConfigFile)
	}

	c := &Config{}

	if full {
		c = NewConfig()
	}

	data, err := readConfig(nil, files...)
	if err != nil {
		return "", err
	}

	err = c.merge(data)
	if err != nil {
		return "", err
	}

	err = c.readGlobals()
	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(c)
	return string(bytes), err
}

func (c *Config) configureConsole() error {
	if console, ok := c.SystemContainers[CONSOLE_CONTAINER]; ok {
		if c.Console.Persistent {
			console.Labels = append(console.Labels, REMOVE+"=false")
		} else {
			console.Labels = append(console.Labels, REMOVE+"=true")
		}
	}

	return nil
}

func (c *Config) readGlobals() error {
	return util.ShortCircuit(
		c.readCmdline,
		c.readArgs,
		c.configureConsole,
	)
}

func (c *Config) Reload() error {
	return util.ShortCircuit(
		c.readFiles,
		c.readGlobals,
	)
}

func (c *Config) Get(key string) (interface{}, error) {
	data := make(map[interface{}]interface{})
	err := util.Convert(c, &data)
	if err != nil {
		return nil, err
	}

	return getOrSetVal(key, data, nil), nil
}

func (c *Config) Set(key string, value interface{}) error {
	data, err := readSavedConfig(nil)
	if err != nil {
		return err
	}

	getOrSetVal(key, data, value)

	err = saveToDisk(data)
	if err != nil {
		return err
	}

	return c.Reload()
}
