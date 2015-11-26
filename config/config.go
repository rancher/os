package config

import (
	log "github.com/Sirupsen/logrus"
	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/util"
)

func (c *CloudConfig) Import(bytes []byte) (*CloudConfig, error) {
	data, err := readConfig(bytes, false, CloudConfigPrivateFile)
	if err != nil {
		return c, err
	}

	cfg := NewConfig()
	if err := util.Convert(data, cfg); err != nil {
		return c, err
	}

	return cfg, nil
}

func (c *CloudConfig) MergeBytes(bytes []byte) (*CloudConfig, error) {
	data, err := readConfig(bytes, false)
	if err != nil {
		return c, err
	}
	return c.Merge(data)
}

func (c *CloudConfig) Merge(values map[interface{}]interface{}) (*CloudConfig, error) {
	t := *c
	if err := util.Convert(values, &t); err != nil {
		return c, err
	}
	return &t, nil
}

func Dump(boot, private, full bool) (string, error) {
	var cfg *CloudConfig
	var err error

	if full {
		cfg, err = LoadConfig()
	} else {
		files := []string{CloudConfigBootFile, CloudConfigPrivateFile, CloudConfigFile}
		if !private {
			files = util.FilterStrings(files, func(x string) bool { return x != CloudConfigPrivateFile })
		}
		if !boot {
			files = util.FilterStrings(files, func(x string) bool { return x != CloudConfigBootFile })
		}
		cfg, err = ChainCfgFuncs(nil,
			func(_ *CloudConfig) (*CloudConfig, error) { return ReadConfig(nil, true, files...) },
			amendNils,
		)
	}

	if err != nil {
		return "", err
	}

	bytes, err := yaml.Marshal(*cfg)
	return string(bytes), err
}

func (c *CloudConfig) Get(key string) (interface{}, error) {
	data := map[interface{}]interface{}{}
	if err := util.Convert(c, &data); err != nil {
		return nil, err
	}

	v, _ := getOrSetVal(key, data, nil)
	return v, nil
}

func (c *CloudConfig) Set(key string, value interface{}) (*CloudConfig, error) {
	data := map[interface{}]interface{}{}
	if err := util.Convert(c, &data); err != nil {
		return c, err
	}

	_, data = getOrSetVal(key, data, value)

	return c.Merge(data)
}

func (c *CloudConfig) Save() error {
	files := append([]string{OsConfigFile}, CloudConfigDirFiles()...)
	files = util.FilterStrings(files, func(x string) bool { return x != CloudConfigPrivateFile })
	exCfg, err := ChainCfgFuncs(nil,
		func(_ *CloudConfig) (*CloudConfig, error) {
			return ReadConfig(nil, true, files...)
		},
		readCmdline,
		amendNils)
	if err != nil {
		return err
	}
	exData := map[interface{}]interface{}{}
	if err := util.Convert(exCfg, &exData); err != nil {
		return err
	}

	data := map[interface{}]interface{}{}
	if err := util.Convert(c, &data); err != nil {
		return err
	}

	data = util.MapsDifference(data, exData)
	log.WithFields(log.Fields{"diff": data}).Debug("The diff we're about to save")
	if err := saveToDisk(data); err != nil {
		return err
	}
	return nil
}
