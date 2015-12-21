package config

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/util"
)

var osConfig *CloudConfig

func NewConfig() *CloudConfig {
	if osConfig == nil {
		osConfig, _ = ReadConfig(nil, true, OsConfigFile)
	}
	newCfg := *osConfig
	return &newCfg
}

func ReadConfig(bytes []byte, substituteMetadataVars bool, files ...string) (*CloudConfig, error) {
	if data, err := readConfig(bytes, substituteMetadataVars, files...); err == nil {
		c := &CloudConfig{}
		if err := util.Convert(data, c); err != nil {
			return nil, err
		}
		c, _ = amendNils(c)
		c, _ = amendContainerNames(c)
		return c, nil
	} else {
		return nil, err
	}
}

func LoadConfig() (*CloudConfig, error) {
	cfg, err := ChainCfgFuncs(NewConfig(),
		readFilesAndMetadata,
		readCmdline,
		amendNils,
		amendContainerNames)
	if err != nil {
		log.WithFields(log.Fields{"cfg": cfg, "err": err}).Error("Failed to load config")
		return nil, err
	}

	log.Debug("Merging cloud-config from meta-data and user-data")
	cfg = mergeMetadata(cfg, readMetadata())

	if cfg.Rancher.Debug {
		log.SetLevel(log.DebugLevel)
		if !util.Contains(cfg.Rancher.Docker.Args, "-D") {
			cfg.Rancher.Docker.Args = append(cfg.Rancher.Docker.Args, "-D")
		}
		if !util.Contains(cfg.Rancher.SystemDocker.Args, "-D") {
			cfg.Rancher.SystemDocker.Args = append(cfg.Rancher.SystemDocker.Args, "-D")
		}
	}

	return cfg, nil
}

func CloudConfigDirFiles() []string {
	files, err := util.DirLs(CloudConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			// do nothing
			log.Debugf("%s does not exist", CloudConfigDir)
		} else {
			log.Errorf("Failed to read %s: %v", CloudConfigDir, err)
		}
		return []string{}
	}

	files = util.Filter(files, func(x interface{}) bool {
		f := x.(os.FileInfo)
		if f.IsDir() || strings.HasPrefix(f.Name(), ".") {
			return false
		}
		return true
	})

	return util.ToStrings(util.Map(files, func(x interface{}) interface{} {
		return path.Join(CloudConfigDir, x.(os.FileInfo).Name())
	}))
}

// mergeMetadata merges certain options from md (meta-data from the datasource)
// onto cc (a CloudConfig derived from user-data), if they are not already set
// on cc (i.e. user-data always takes precedence)
func mergeMetadata(cc *CloudConfig, md datasource.Metadata) *CloudConfig {
	if cc == nil {
		return cc
	}
	out := cc
	dirty := false

	if md.Hostname != "" {
		if out.Hostname != "" {
			log.Debugf("Warning: user-data hostname (%s) overrides metadata hostname (%s)\n", out.Hostname, md.Hostname)
		} else {
			out = &(*cc)
			dirty = true
			out.Hostname = md.Hostname
		}
	}
	for _, key := range md.SSHPublicKeys {
		if !dirty {
			out = &(*cc)
			dirty = true
		}
		out.SSHAuthorizedKeys = append(out.SSHAuthorizedKeys, key)
	}
	return out
}

func readMetadata() datasource.Metadata {
	metadata := datasource.Metadata{}
	if metaDataBytes, err := ioutil.ReadFile(MetaDataFile); err == nil {
		yaml.Unmarshal(metaDataBytes, &metadata)
	}
	return metadata
}

func readFilesAndMetadata(c *CloudConfig) (*CloudConfig, error) {
	files := append(CloudConfigDirFiles(), CloudConfigFile)
	data, err := readConfig(nil, true, files...)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "files": files}).Error("Error reading config files")
		return c, err
	}

	t, err := c.Merge(data)
	if err != nil {
		log.WithFields(log.Fields{"cfg": c, "data": data, "err": err}).Error("Error merging config data")
		return c, err
	}

	return t, nil
}

func readCmdline(c *CloudConfig) (*CloudConfig, error) {
	log.Debug("Reading config cmdline")
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read kernel params")
		return c, err
	}

	if len(cmdLine) == 0 {
		return c, nil
	}

	log.Debugf("Config cmdline %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))

	t, err := c.Merge(cmdLineObj)
	if err != nil {
		log.WithFields(log.Fields{"cfg": c, "cmdLine": cmdLine, "data": cmdLineObj, "err": err}).Warn("Error adding kernel params to config")
	}
	return t, nil
}

func amendNils(c *CloudConfig) (*CloudConfig, error) {
	t := *c
	if t.Rancher.Environment == nil {
		t.Rancher.Environment = map[string]string{}
	}
	if t.Rancher.Autoformat == nil {
		t.Rancher.Autoformat = map[string]*project.ServiceConfig{}
	}
	if t.Rancher.BootstrapContainers == nil {
		t.Rancher.BootstrapContainers = map[string]*project.ServiceConfig{}
	}
	if t.Rancher.Services == nil {
		t.Rancher.Services = map[string]*project.ServiceConfig{}
	}
	if t.Rancher.ServicesInclude == nil {
		t.Rancher.ServicesInclude = map[string]bool{}
	}
	return &t, nil
}

func amendContainerNames(c *CloudConfig) (*CloudConfig, error) {
	for _, scm := range []map[string]*project.ServiceConfig{
		c.Rancher.Autoformat,
		c.Rancher.BootstrapContainers,
		c.Rancher.Services,
	} {
		for k, v := range scm {
			v.ContainerName = k
		}
	}
	return c, nil
}

func writeToFile(data interface{}, filename string) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, content, 400)
}

func saveToDisk(data map[interface{}]interface{}) error {
	private, config := filterDottedKeys(data, []string{
		"rancher.ssh",
		"rancher.docker.ca_key",
		"rancher.docker.ca_cert",
		"rancher.docker.server_key",
		"rancher.docker.server_cert",
	})

	err := writeToFile(config, CloudConfigFile)
	if err != nil {
		return err
	}

	return writeToFile(private, CloudConfigPrivateFile)
}

func readConfig(bytes []byte, substituteMetadataVars bool, files ...string) (map[interface{}]interface{}, error) {
	// You can't just overlay yaml bytes on to maps, it won't merge, but instead
	// just override the keys and not merge the map values.
	left := make(map[interface{}]interface{})
	metadata := readMetadata()
	for _, file := range files {
		content, err := readConfigFile(file)
		if err != nil {
			return nil, err
		}
		if len(content) == 0 {
			continue
		}
		if substituteMetadataVars {
			content = substituteVars(content, metadata)
		}

		right := make(map[interface{}]interface{})
		err = yaml.Unmarshal(content, &right)
		if err != nil {
			return nil, err
		}

		left = util.MapsUnion(left, right)
	}

	if bytes != nil && len(bytes) > 0 {
		right := make(map[interface{}]interface{})
		if substituteMetadataVars {
			bytes = substituteVars(bytes, metadata)
		}
		if err := yaml.Unmarshal(bytes, &right); err != nil {
			return nil, err
		}

		left = util.MapsUnion(left, right)
	}

	return left, nil
}

func readConfigFile(file string) ([]byte, error) {
	content, err := ioutil.ReadFile(file)

	if err != nil {
		if os.IsNotExist(err) {
			err = nil
			content = []byte{}
		} else {
			return nil, err
		}
	}

	return content, err
}

func substituteVars(userDataBytes []byte, metadata datasource.Metadata) []byte {
	env := initialize.NewEnvironment("", "", "", "", metadata)
	userData := env.Apply(string(userDataBytes))

	return []byte(userData)
}
