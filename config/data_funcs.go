package config

import (
	"github.com/burmilla/os/pkg/log"
	"github.com/burmilla/os/pkg/util"
)

type CfgFunc func(*CloudConfig) (*CloudConfig, error)
type CfgFuncData struct {
	Name string
	Func CfgFunc
}

type CfgFuncs []CfgFuncData

func ChainCfgFuncs(cfg *CloudConfig, cfgFuncs CfgFuncs) (*CloudConfig, error) {
	len := len(cfgFuncs)
	for c, d := range cfgFuncs {
		i := c + 1
		name := d.Name
		cfgFunc := d.Func
		if cfg == nil {
			log.Infof("[%d/%d] Starting %s WITH NIL cfg", i, len, name)
		} else {
			log.Infof("[%d/%d] Starting %s", i, len, name)
		}
		var err error
		if cfg, err = cfgFunc(cfg); err != nil {
			log.Errorf("Failed [%d/%d] %s: %v", i, len, name, err)
			return cfg, err
		}
		log.Debugf("[%d/%d] Done %s", i, len, name)
	}
	return cfg, nil
}

func filterKey(data map[interface{}]interface{}, key []string) (filtered, rest map[interface{}]interface{}) {
	if len(key) == 0 {
		return data, map[interface{}]interface{}{}
	}

	filtered = map[interface{}]interface{}{}
	rest = util.MapCopy(data)

	k := key[0]
	if d, ok := data[k]; ok {
		switch d := d.(type) {

		case map[interface{}]interface{}:
			f, r := filterKey(d, key[1:])

			if len(f) != 0 {
				filtered[k] = f
			}

			if len(r) != 0 {
				rest[k] = r
			} else {
				delete(rest, k)
			}

		default:
			filtered[k] = d
			delete(rest, k)
		}

	}

	return
}
