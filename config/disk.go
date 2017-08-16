package config

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/docker/engine-api/types"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/initialize"
	"github.com/rancher/os/config/cmdline"
	"github.com/rancher/os/log"
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

func loadRawDiskConfig(dirPrefix string, full bool) map[interface{}]interface{} {
	var rawCfg map[interface{}]interface{}
	if full {
		rawCfg, _ = readConfigs(nil, true, false, OsConfigFile, OemConfigFile)
	}

	files := CloudConfigDirFiles(dirPrefix)
	files = append(files, path.Join(dirPrefix, CloudConfigFile))
	additionalCfgs, _ := readConfigs(nil, true, false, files...)

	return util.Merge(rawCfg, additionalCfgs)
}

func loadRawConfig(dirPrefix string, full bool) map[interface{}]interface{} {
	rawCfg := loadRawDiskConfig(dirPrefix, full)
	procCmdline, err := cmdline.Read(false)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read kernel params")
	}
	rawCfg = util.Merge(rawCfg, procCmdline)
	rawCfg = util.Merge(rawCfg, readElidedCmdline(rawCfg))
	rawCfg = applyDebugFlags(rawCfg)
	return mergeMetadata(rawCfg, readMetadata())
}

func LoadConfig() *CloudConfig {
	cfg := LoadConfigWithPrefix("")

	if cfg.Rancher.Debug {
		log.SetDefaultLevel(log.DebugLevel)
	} else {
		log.SetDefaultLevel(log.InfoLevel)
	}

	return cfg
}

func LoadConfigWithPrefix(dirPrefix string) *CloudConfig {
	rawCfg := loadRawConfig(dirPrefix, true)

	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		log.Errorf("EXITING: Failed to parse configuration: %s", err)
		log.Debugf("Bad cfg:\n%v\n", rawCfg)
		// no point returning {}, it'll just sit there broken
		// TODO: print some context around what failed..
		validationErrors, err := ValidateRawCfg(rawCfg)
		if err != nil {
			log.Fatal(err)
		}
		for _, validationError := range validationErrors.Errors() {
			log.Error(validationError)
		}
		// TODO: I'd love to panic & recover(), for issues on boot, but it doesn't work yet
		os.Exit(-1)
		return &CloudConfig{}
	}
	cfg = amendNils(cfg)
	cfg = amendContainerNames(cfg)
	return cfg
}

func Insert(m interface{}, args ...interface{}) interface{} {
	// TODO: move to util.go
	if len(args)%2 != 0 {
		panic("must have pairs of keys and values")
	}
	mv := reflect.ValueOf(m)
	if mv.IsNil() {
		mv = reflect.MakeMap(mv.Type())
	}
	for i := 0; i < len(args); i += 2 {
		mv.SetMapIndex(reflect.ValueOf(args[i]), reflect.ValueOf(args[i+1]))
	}
	return mv.Interface()
}

func SaveInitCmdline(cmdLineArgs string) {
	elidedCfg := cmdline.Parse(cmdLineArgs, false)

	env := Insert(make(map[interface{}]interface{}), interface{}("EXTRA_CMDLINE"), interface{}(cmdLineArgs))
	rancher := Insert(make(map[interface{}]interface{}), interface{}("environment"), env)
	newCfg := Insert(elidedCfg, interface{}("rancher"), rancher)
	// make it easy for readElidedCmdline(rawCfg)
	newCfg = Insert(newCfg, interface{}("EXTRA_CMDLINE"), interface{}(cmdLineArgs))

	if err := WriteToFile(newCfg, CloudConfigInitFile); err != nil {
		log.Errorf("Failed to write init-cmdline config: %s", err)
	}
}

func CloudConfigDirFiles(dirPrefix string) []string {
	cloudConfigDir := path.Join(dirPrefix, CloudConfigDir)

	files, err := ioutil.ReadDir(cloudConfigDir)
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
			finalFiles = append(finalFiles, path.Join(cloudConfigDir, file.Name()))
		}
	}

	return finalFiles
}

func applyDebugFlags(rawCfg map[interface{}]interface{}) map[interface{}]interface{} {
	cfg := &CloudConfig{}
	if err := util.Convert(rawCfg, cfg); err != nil {
		return rawCfg
	}

	if !cfg.Rancher.Debug {
		return rawCfg
	}

	log.SetLevel(log.DebugLevel)
	_, rawCfg = cmdline.GetOrSetVal("rancher.docker.debug", rawCfg, true)
	_, rawCfg = cmdline.GetOrSetVal("rancher.system_docker.debug", rawCfg, true)
	_, rawCfg = cmdline.GetOrSetVal("rancher.bootstrap_docker.debug", rawCfg, true)
	_, rawCfg = cmdline.GetOrSetVal("rancher.log", rawCfg, true)

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

	finalKeys, _ := out["ssh_authorized_keys"].([]interface{})
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

func readElidedCmdline(rawCfg map[interface{}]interface{}) map[interface{}]interface{} {

	for k, v := range rawCfg {
		if key, _ := k.(string); key == "EXTRA_CMDLINE" {
			if val, ok := v.(string); ok {
				cmdLineObj := cmdline.Parse(strings.TrimSpace(util.UnescapeKernelParams(string(val))), false)

				return cmdLineObj
			}
		}
	}
	return nil
}

func amendNils(c *CloudConfig) *CloudConfig {
	t := *c
	if t.Rancher.Environment == nil {
		t.Rancher.Environment = map[string]string{}
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
	if t.Rancher.RegistryAuths == nil {
		t.Rancher.RegistryAuths = map[string]types.AuthConfig{}
	}
	if t.Rancher.Sysctl == nil {
		t.Rancher.Sysctl = map[string]string{}
	}
	return &t
}

func amendContainerNames(c *CloudConfig) *CloudConfig {
	for _, scm := range []map[string]*composeConfig.ServiceConfigV1{
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

	if err := os.MkdirAll(filepath.Dir(filename), os.ModeDir|0700); err != nil {
		return err
	}

	return util.WriteFileAtomic(filename, content, 400)
}

func readConfigs(bytes []byte, substituteMetadataVars, returnErr bool, files ...string) (map[interface{}]interface{}, error) {
	// You can't just overlay yaml bytes on to maps, it won't merge, but instead
	// just override the keys and not merge the map values.
	left := make(map[interface{}]interface{})
	metadata := readMetadata()
	for _, file := range files {
		//os.Stderr.WriteString(fmt.Sprintf("READCONFIGS(%s)", file))
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
	// TODO: I think this currently does nothing - its hardcoded for COREOS env..
	env := initialize.NewEnvironment("", "", "", "", metadata)
	userData := env.Apply(string(userDataBytes))

	return []byte(userData)
}
