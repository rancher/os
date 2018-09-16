package control

import (
	"log"
	"os"
	"os/exec"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/util"
)

func envAction(c *cli.Context) error {
	cfg := config.LoadConfig()

	args := c.Args()
	if len(args) == 0 {
		return nil
	}
	osEnv := os.Environ()

	envMap := make(map[string]string, len(cfg.Rancher.Environment)+len(osEnv))
	for k, v := range cfg.Rancher.Environment {
		envMap[k] = v
	}
	for k, v := range util.KVPairs2Map(osEnv) {
		envMap[k] = v
	}

	if cmd, err := exec.LookPath(args[0]); err != nil {
		log.Fatal(err)
	} else {
		args[0] = cmd
	}
	if err := syscall.Exec(args[0], args, util.Map2KVPairs(envMap)); err != nil {
		log.Fatal(err)
	}

	return nil
}
