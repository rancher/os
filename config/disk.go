package config

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/util"
)

func NewConfig() map[interface{}]interface{} {
	osConfig, _ := readConfig(nil, true, OsConfigFile, OemConfigFile)
	return osConfig
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

func LoadRawConfig(full bool) (map[interface{}]interface{}, error) {
	var base map[interface{}]interface{}
	if full {
		base = NewConfig()
	}
	user, err := readConfigs()
	if err != nil {
		return nil, err
	}
	cmdline, err := readCmdline()
	if err != nil {
		return nil, err
	}
	merged := util.Merge(base, util.Merge(user, cmdline))
	merged, err = applyDebugFlags(merged)
	if err != nil {
		return nil, err
	}
	return mergeMetadata(merged, readMetadata()), nil
}

func LoadConfig() (*CloudConfig, error) {
	rawCfg, err := LoadRawConfig(true)
	if err != nil {
		return nil, err
	}

	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		return nil, err
	}
	cfg, err = amendNils(cfg)
	if err != nil {
		return nil, err
	}
	cfg, err = amendContainerNames(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func CloudConfigDirFiles() []string {
	files, err := ioutil.ReadDir(CloudConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			// do nothing
			log.Debugf("%s does not exist", CloudConfigDir)
		} else {
			log.Errorf("Failed to read %s: %v", CloudConfigDir, err)
		}
		return []string{}
	}

	var finalFiles []string
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			finalFiles = append(finalFiles, path.Join(CloudConfigDir, file.Name()))
		}
	}

	return finalFiles
}

func applyDebugFlags(rawCfg map[interface{}]interface{}) (map[interface{}]interface{}, error) {
	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		return nil, err
	}

	if cfg.Rancher.Debug {
		log.SetLevel(log.DebugLevel)
		if !util.Contains(cfg.Rancher.Docker.Args, "-D") {
			cfg.Rancher.Docker.Args = append(cfg.Rancher.Docker.Args, "-D")
		}
		if !util.Contains(cfg.Rancher.SystemDocker.Args, "-D") {
			cfg.Rancher.SystemDocker.Args = append(cfg.Rancher.SystemDocker.Args, "-D")
		}
	} else {
		if util.Contains(cfg.Rancher.Docker.Args, "-D") {
			cfg.Rancher.Docker.Args = util.RemoveString(cfg.Rancher.Docker.Args, "-D")
		}
		if util.Contains(cfg.Rancher.SystemDocker.Args, "-D") {
			cfg.Rancher.SystemDocker.Args = util.RemoveString(cfg.Rancher.SystemDocker.Args, "-D")
		}
	}

	_, rawCfg = getOrSetVal("rancher.docker.args", rawCfg, cfg.Rancher.Docker.Args)
	_, rawCfg = getOrSetVal("rancher.system_docker.args", rawCfg, cfg.Rancher.SystemDocker.Args)
	return rawCfg, nil
}

// mergeMetadata merges certain options from md (meta-data from the datasource)
// onto cc (a CloudConfig derived from user-data), if they are not already set
// on cc (i.e. user-data always takes precedence)
func mergeMetadata(rawCfg map[interface{}]interface{}, md datasource.Metadata) map[interface{}]interface{} {
	if rawCfg == nil {
		return nil
	}
	out := util.MapCopy(rawCfg)

	outHostname, ok := out["hostname"]
	if !ok {
		outHostname = ""
	}

	if md.Hostname != "" {
		if outHostname != "" {
			log.Debugf("Warning: user-data hostname (%s) overrides metadata hostname (%s)\n", outHostname, md.Hostname)
		} else {
			out["hostname"] = md.Hostname
		}
	}

	// Sort SSH keys by key name
	keys := []string{}
	for k := range md.SSHPublicKeys {
		keys = append(keys, k)
	}

	sort.Sort(sort.StringSlice(keys))

	currentKeys, ok := out["ssh_authorized_keys"]
	if !ok {
		return out
	}

	finalKeys := currentKeys.([]interface{})
	for _, k := range keys {
		finalKeys = append(finalKeys, md.SSHPublicKeys[k])
	}

	out["ssh_authorized_keys"] = finalKeys

	return out
}

func readMetadata() datasource.Metadata {
	metadata := datasource.Metadata{}
	if metaDataBytes, err := ioutil.ReadFile(MetaDataFile); err == nil {
		yaml.Unmarshal(metaDataBytes, &metadata)
	}
	return metadata
}

func readConfigs() (map[interface{}]interface{}, error) {
	files := append(CloudConfigDirFiles(), CloudConfigFile)
	data, err := readConfig(nil, true, files...)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func readCmdline() (map[interface{}]interface{}, error) {
	log.Debug("Reading config cmdline")
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read kernel params")
		return nil, err
	}

	if len(cmdLine) == 0 {
		return nil, nil
	}

	log.Debugf("Config cmdline %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))

	return cmdLineObj, nil
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

func WriteToFile(data interface{}, filename string) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, content, 400)
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

		left = util.Merge(left, right)
	}

	if bytes != nil && len(bytes) > 0 {
		right := make(map[interface{}]interface{})
		if substituteMetadataVars {
			bytes = substituteVars(bytes, metadata)
		}
		if err := yaml.Unmarshal(bytes, &right); err != nil {
			return nil, err
		}

		left = util.Merge(left, right)
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
