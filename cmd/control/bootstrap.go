package control

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codegangsta/cli"

	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"
)

func bootstrapAction(c *cli.Context) error {
	log.Debugf("bootstrapAction")
	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}

	log.Debugf("bootstrapAction: loadingConfig")
	cfg := config.LoadConfig()

	log.Debugf("bootstrapAction: MdadmScan(%v)", cfg.Rancher.State.MdadmScan)
	if cfg.Rancher.State.MdadmScan {
		if err := mdadmScan(); err != nil {
			log.Errorf("Failed to run mdadm scan: %v", err)
		}
	}

	stateScript := cfg.Rancher.State.Script
	log.Debugf("bootstrapAction: stateScript(%v)", stateScript)
	if stateScript != "" {
		if err := runStateScript(stateScript); err != nil {
			log.Errorf("Failed to run state script: %v", err)
		}
	}

	log.Debugf("bootstrapAction: RunCommandSequence(%v)", cfg.Bootcmd)
	util.RunCommandSequence(cfg.Bootcmd)

	if cfg.Rancher.State.Dev != "" && cfg.Rancher.State.Wait {
		waitForRoot(cfg)
	}

	autoformatDevices := cfg.Rancher.State.Autoformat
	log.Debugf("bootstrapAction: Autoformat(%v)", cfg.Rancher.State.Autoformat)
	if len(autoformatDevices) > 0 {
		if err := autoformat(autoformatDevices); err != nil {
			log.Errorf("Failed to run autoformat: %v", err)
		}
	}

	log.Debugf("bootstrapAction: udev settle2")
	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}

	return nil
}

func mdadmScan() error {
	cmd := exec.Command("mdadm", "--assemble", "--scan")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runStateScript(script string) error {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	if _, err := f.WriteString(script); err != nil {
		return err
	}
	if err := f.Chmod(os.ModePerm); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return util.RunScript(f.Name())
}

func waitForRoot(cfg *config.CloudConfig) {
	var dev string
	for i := 0; i < 30; i++ {
		dev = util.ResolveDevice(cfg.Rancher.State.Dev)
		if dev != "" {
			break
		}
		time.Sleep(time.Millisecond * 1000)
	}
	if dev == "" {
		return
	}
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(dev); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func autoformat(autoformatDevices []string) error {
	cmd := exec.Command("/usr/sbin/auto-format.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		"AUTOFORMAT=" + strings.Join(autoformatDevices, " "),
	}
	return cmd.Run()
}
