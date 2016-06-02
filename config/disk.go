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
	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/util"
)

func ReadConfig(bytes []byte, substituteMetadataVars bool, files ...string) (*CloudConfig, error) {
	data, err := readConfigs(bytes, substituteMetadataVars, true, files...)
	if err != nil {
		return nil, err
	}

	c := &CloudConfig{}
	if err := util.Convert(data, c); err != nil {
		return nil, err
	}
	c = amendNils(c)
	c = amendContainerNames(c)
	return c, nil
}

func loadRawDiskConfig(full bool) map[interface{}]interface{} {
	var rawCfg map[interface{}]interface{}
	if full {
		rawCfg, _ = readConfigs(nil, true, false, OsConfigFile, OemConfigFile)
	}

	files := append(CloudConfigDirFiles(), CloudConfigFile)
	additionalCfgs, _ := readConfigs(nil, true, false, files...)

	return util.Merge(rawCfg, additionalCfgs)
}

func loadRawConfig() map[interface{}]interface{} {
	rawCfg := loadRawDiskConfig(true)
	rawCfg = util.Merge(rawCfg, readCmdline())
	rawCfg = applyDebugFlags(rawCfg)
	return mergeMetadata(rawCfg, readMetadata())
}

func LoadConfig() *CloudConfig {
	rawCfg := loadRawConfig()

	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		log.Errorf("Failed to parse configuration: %s", err)
		return &CloudConfig{}
	}
	cfg = amendNils(cfg)
	cfg = amendContainerNames(cfg)
	return cfg
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

func applyDebugFlags(rawCfg map[interface{}]interface{}) map[interface{}]interface{} {
	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		return rawCfg
	}

	if cfg.Rancher.Debug {
		log.SetLevel(log.DebugLevel)
		if !util.Contains(cfg.Rancher.Docker.Args, "-D") {
			cfg.Rancher.Docker.Args = append(cfg.Rancher.Docker.Args, "-D")
		}
		if !util.Contains(cfg.Rancher.SystemDocker.Args, "-D") {
			cfg.Rancher.SystemDocker.Args = append(cfg.Rancher.SystemDocker.Args, "-D")
		}
	}

	_, rawCfg = getOrSetVal("rancher.docker.args", rawCfg, cfg.Rancher.Docker.Args)
	_, rawCfg = getOrSetVal("rancher.system_docker.args", rawCfg, cfg.Rancher.SystemDocker.Args)
	return rawCfg
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

func readCmdline() map[interface{}]interface{} {
	log.Debug("Reading config cmdline")
	cmdLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read kernel params")
		return nil
	}

	if len(cmdLine) == 0 {
		return nil
	}

	log.Debugf("Config cmdline %s", cmdLine)

	cmdLineObj := parseCmdline(strings.TrimSpace(string(cmdLine)))

	return cmdLineObj
}

func amendNils(c *CloudConfig) *CloudConfig {
	t := *c
	if t.Rancher.Environment == nil {
		t.Rancher.Environment = map[string]string{}
	}
	if t.Rancher.Autoformat == nil {
		t.Rancher.Autoformat = map[string]*composeConfig.ServiceConfigV1{}
	}
	if t.Rancher.BootstrapContainers == nil {
		t.Rancher.BootstrapContainers = map[string]*composeConfig.ServiceConfigV1{}
	}
	if t.Rancher.Services == nil {
		t.Rancher.Services = map[string]*composeConfig.ServiceConfigV1{}
	}
	if t.Rancher.ServicesInclude == nil {
		t.Rancher.ServicesInclude = map[string]bool{}
	}
	return &t
}

func amendContainerNames(c *CloudConfig) *CloudConfig {
	for _, scm := range []map[string]*composeConfig.ServiceConfigV1{
		c.Rancher.Autoformat,
		c.Rancher.BootstrapContainers,
		c.Rancher.Services,
	} {
		for k, v := range scm {
			v.ContainerName = k
		}
	}
	return c
}

func WriteToFile(data interface{}, filename string) error {
	content, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, content, 400)
}

func readConfigs(bytes []byte, substituteMetadataVars, returnErr bool, files ...string) (map[interface{}]interface{}, error) {
	// You can't just overlay yaml bytes on to maps, it won't merge, but instead
	// just override the keys and not merge the map values.
	left := make(map[interface{}]interface{})
	metadata := readMetadata()
	for _, file := range files {
		content, err := readConfigFile(file)
		if err != nil {
			if returnErr {
				return nil, err
			}
			log.Errorf("Failed to read config file %s: %s", file, err)
			continue
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
			if returnErr {
				return nil, err
			}
			log.Errorf("Failed to parse config file %s: %s", file, err)
			continue
		}

		// Verify there are no issues converting to CloudConfig
		c := &CloudConfig{}
		if err := util.Convert(right, c); err != nil {
			if returnErr {
				return nil, err
			}
			log.Errorf("Failed to parse config file %s: %s", file, err)
			continue
		}

		left = util.Merge(left, right)
	}

	if bytes == nil || len(bytes) == 0 {
		return left, nil
	}

	right := make(map[interface{}]interface{})
	if substituteMetadataVars {
		bytes = substituteVars(bytes, metadata)
	}

	if err := yaml.Unmarshal(bytes, &right); err != nil {
		if returnErr {
			return nil, err
		}
		log.Errorf("Failed to parse bytes: %s", err)
		return left, nil
	}

	c := &CloudConfig{}
	if err := util.Convert(right, c); err != nil {
		if returnErr {
			return nil, err
		}
		log.Errorf("Failed to parse bytes: %s", err)
		return left, nil
	}

	left = util.Merge(left, right)
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
