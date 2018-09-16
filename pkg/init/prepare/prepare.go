package prepare

import (
	"os"
	"strings"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/dfs"
	"github.com/rancher/os/pkg/log"
)

var (
	mountConfig = dfs.Config{
		CgroupHierarchy: map[string]string{
			"cpu":      "cpu",
			"cpuacct":  "cpu",
			"net_cls":  "net_cls",
			"net_prio": "net_cls",
		},
	}
)

func FS(c *config.CloudConfig) (*config.CloudConfig, error) {
	return c, dfs.PrepareFs(&mountConfig)
}

func SaveCmdline(c *config.CloudConfig) (*config.CloudConfig, error) {
	// the Kernel Patch added for RancherOS passes `--` (only) elided kernel boot params to the init process
	cmdLineArgs := strings.Join(os.Args, " ")
	config.SaveInitCmdline(cmdLineArgs)

	cfg := config.LoadConfig()
	log.Debugf("Cmdline debug = %t", cfg.Rancher.Debug)
	if cfg.Rancher.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	return cfg, nil
}
