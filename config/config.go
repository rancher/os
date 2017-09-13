package config

import (
	"io/ioutil"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/config/cmdline"
	"github.com/rancher/os/util"
)

const Banner = `
               ,        , ______                 _                 _____ _____TM
  ,------------|'------'| | ___ \\               | |               /  _  /  ___|
 / .           '-'    |-  | |_/ /__ _ _ __   ___| |__   ___ _ __  | | | \\ '--.
 \\/|             |    |   |    // _' | '_ \\ / __| '_ \\ / _ \\ '__' | | | |'--. \\
   |   .________.'----'   | |\\ \\ (_| | | | | (__| | | |  __/ |    | \\_/ /\\__/ /
   |   |        |   |     \\_| \\_\\__,_|_| |_|\\___|_| |_|\\___|_|     \\___/\\____/
   \\___/        \\___/     \s \r

         RancherOS \v \n \l
         `

func Merge(bytes []byte) error {
	data, err := readConfigs(bytes, false, true)
	if err != nil {
		return err
	}
	existing, err := readConfigs(nil, false, true, CloudConfigFile)
	if err != nil {
		return err
	}
	return WriteToFile(util.Merge(existing, data), CloudConfigFile)
}

func Export(private, full bool) (string, error) {
	rawCfg := loadRawConfig("", full)
	if !private {
		rawCfg = filterPrivateKeys(rawCfg)
	}

	bytes, err := yaml.Marshal(rawCfg)
	return "#cloud-config\n" + string(bytes), err
}
func filterPrivateKeys(data map[interface{}]interface{}) map[interface{}]interface{} {
	for _, privateKey := range PrivateKeys {
		_, data = filterKey(data, strings.Split(privateKey, "."))
	}

	return data
}

func Get(key string) (interface{}, error) {
	cfg := LoadConfig()

	data := map[interface{}]interface{}{}
	if err := util.ConvertIgnoreOmitEmpty(cfg, &data); err != nil {
		return nil, err
	}

	v, _ := cmdline.GetOrSetVal(key, data, nil)
	return v, nil
}

func Set(key string, value interface{}) error {
	existing, err := readConfigs(nil, false, true, CloudConfigFile)
	if err != nil {
		return err
	}

	_, modified := cmdline.GetOrSetVal(key, existing, value)

	c := &CloudConfig{}
	if err = util.Convert(modified, c); err != nil {
		return err
	}

	return WriteToFile(modified, CloudConfigFile)
}

func GetKernelVersion() string {
	b, err := ioutil.ReadFile("/proc/version")
	if err != nil {
		return ""
	}
	elem := strings.Split(string(b), " ")
	return elem[2]
}
