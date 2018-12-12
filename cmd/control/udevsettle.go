package control

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"

	"github.com/codegangsta/cli"
)

func udevSettleAction(c *cli.Context) {
	if err := extraRules(); err != nil {
		log.Error(err)
	}

	if err := UdevSettle(); err != nil {
		log.Fatal(err)
	}
}

func extraRules() error {
	cfg := config.LoadConfig()
	if len(cfg.Rancher.Network.ModemNetworks) > 0 {
		rules, err := ioutil.ReadDir(config.UdevRulesExtrasDir)
		if err != nil {
			return err
		}
		for _, r := range rules {
			if r.IsDir() || filepath.Ext(r.Name()) != ".rules" {
				continue
			}
			err := os.Symlink(filepath.Join(config.UdevRulesExtrasDir, r.Name()), filepath.Join(config.UdevRulesDir, r.Name()))
			if err != nil {
				return err
			}
		}
	} else {
		rules, err := ioutil.ReadDir(config.UdevRulesDir)
		if err != nil {
			return err
		}
		for _, r := range rules {
			if r.IsDir() || (filepath.Ext(r.Name()) != ".rules") || (r.Mode()&os.ModeSymlink != 0) {
				continue
			}
			err := os.Remove(filepath.Join(config.UdevRulesDir, r.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func UdevSettle() error {
	cmd := exec.Command("udevd", "--daemon")
	defer exec.Command("killall", "udevd").Run()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("udevadm", "trigger", "--action=add")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("udevadm", "settle")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
