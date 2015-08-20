package config

import (
	"io/ioutil"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/project"
	"github.com/rancherio/os/util"
	"gopkg.in/yaml.v2"
)

func (c *CloudConfig) Import(bytes []byte) error {
	data, err := readConfig(bytes, PrivateConfigFile)
	if err != nil {
		return err
	}

	if err := saveToDisk(data); err != nil {
		return err
	}

	return c.Reload()
}

// This function only sets "non-empty" values
func (c *CloudConfig) SetConfig(newConfig *CloudConfig) error {
	bytes, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}

	return c.Merge(bytes)
}

func (c *CloudConfig) Merge(bytes []byte) error {
	data, err := readConfig(bytes, LocalConfigFile, PrivateConfigFile)
	if err != nil {
		return err
	}

	if err := saveToDisk(data); err != nil {
		return err
	}

	return c.Reload()
}

func LoadConfig() (*CloudConfig, error) {
	cfg := NewConfig()
	if err := cfg.Reload(); err != nil {
		log.WithFields(log.Fields{"cfg": cfg, "err": err}).Error("Failed to reload config")
		return nil, err
	}

	if cfg.Rancher.Debug {
		log.SetLevel(log.DebugLevel)
		if !util.Contains(cfg.Rancher.UserDocker.Args, "-D") {
			cfg.Rancher.UserDocker.Args = append(cfg.Rancher.UserDocker.Args, "-D")
		}
		if !util.Contains(cfg.Rancher.SystemDocker.Args, "-D") {
			cfg.Rancher.SystemDocker.Args = append(cfg.Rancher.SystemDocker.Args, "-D")
		}
	}

	return cfg, nil
}

func (c *CloudConfig) merge(values map[interface{}]interface{}) error {
	t := &CloudConfig{}
	if err := util.Convert(values, t); err != nil {
		return err
	}
	return util.Convert(values, c)
}

func (c *CloudConfig) readFiles() error {
	data, err := readConfig(nil, CloudConfigFile, LocalConfigFile, PrivateConfigFile)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error reading config files")
		return err
	}

	if err := c.merge(data); err != nil {
		log.WithFields(log.Fields{"cfg": c, "data": data, "err": err}).Error("Error merging config data")
		return err
	}

	return nil
}

func (c *CloudConfig) readCmdline() error {
	log.Debug("Reading config cmdline")
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read kernel params")
		return err
	}

	if len(cmdLine) == 0 {
		return nil
	}

	log.Debugf("Config cmdline %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))

	if err := c.merge(cmdLineObj); err != nil {
		log.WithFields(log.Fields{"cfg": c, "cmdLine": cmdLine, "data": cmdLineObj, "err": err}).Warn("Error adding kernel params to config")
	}
	return nil
}

func Dump(private, full bool) (string, error) {
	files := []string{CloudConfigFile, LocalConfigFile}
	if private {
		files = append(files, PrivateConfigFile)
	}

	c := &CloudConfig{}

	if full {
		c = NewConfig()
	}

	data, err := readConfig(nil, files...)
	if err != nil {
		return "", err
	}

	if err := c.merge(data); err != nil {
		return "", err
	}

	if err := c.readCmdline(); err != nil {
		return "", err
	}

	c.amendNils()

	bytes, err := yaml.Marshal(c)
	return string(bytes), err
}

func (c *CloudConfig) amendNils() error {
	if c.Rancher.Environment == nil {
		c.Rancher.Environment = map[string]string{}
	}
	if c.Rancher.Autoformat == nil {
		c.Rancher.Autoformat = map[string]*project.ServiceConfig{}
	}
	if c.Rancher.BootstrapContainers == nil {
		c.Rancher.BootstrapContainers = map[string]*project.ServiceConfig{}
	}
	if c.Rancher.Services == nil {
		c.Rancher.Services = map[string]*project.ServiceConfig{}
	}
	if c.Rancher.ServicesInclude == nil {
		c.Rancher.ServicesInclude = map[string]bool{}
	}
	return nil
}

func (c *CloudConfig) Reload() error {
	return util.ShortCircuit(
		c.readFiles,
		c.readCmdline,
		c.amendNils,
	)
}

func (c *CloudConfig) Get(key string) (interface{}, error) {
	data := make(map[interface{}]interface{})
	err := util.Convert(c, &data)
	if err != nil {
		return nil, err
	}

	return getOrSetVal(key, data, nil), nil
}

func (c *CloudConfig) Set(key string, value interface{}) error {
	data, err := readConfig(nil, LocalConfigFile, PrivateConfigFile)
	if err != nil {
		return err
	}

	getOrSetVal(key, data, value)

	cfg := NewConfig()

	if err := util.Convert(data, cfg); err != nil {
		return err
	}

	if err := saveToDisk(data); err != nil {
		return err
	}

	return c.Reload()
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
