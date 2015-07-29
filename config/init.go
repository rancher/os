package config

import (
	"strings"

	log "github.com/Sirupsen/logrus"
)

type InitFunc func(*CloudConfig) error

func RunInitFuncs(cfg *CloudConfig, initFuncs []InitFunc) error {
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

func FilterGlobalConfig(input []string) []string {
	result := make([]string, 0, len(input))
	for _, value := range input {
		if !strings.HasPrefix(value, "--rancher") {
			result = append(result, value)
		}
	}

	return result
}
