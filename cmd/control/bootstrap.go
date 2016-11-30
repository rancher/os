package control

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/codegangsta/cli"

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

func bootstrapAction(c *cli.Context) error {
	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}

	cfg := config.LoadConfig()

	if cfg.Rancher.State.MdadmScan {
		cmdRun("mdadm", "--assemble", "--scan")
	}

	if cfg.Rancher.State.NetConf || cfg.Rancher.State.Nbd.Host != "" || cfg.Rancher.State.Decrypt {
                if err := cmdStart("netconf"); err != nil {
                        log.Errorf("Failed to start netconf: %v", err)
                }
	}

	if cfg.Rancher.State.Nbd.Host != "" {
		nbd := cfg.Rancher.State.Nbd
		args := []string { nbd.Host }
		if nbd.Port != 0 {
			args = append(args, string(nbd.Port))
		}
		if nbd.Name != "" {
			args = append(args, "-name", nbd.Name)
		}
		if nbd.BlockSize != 0 {
			args = append(args, "-block-size", string(nbd.BlockSize))
		}
                cmdRun("nbd-client", args...)
	}

	stateScript := cfg.Rancher.State.Script
	if stateScript != "" {
		if err := runStateScript(stateScript); err != nil {
			log.Errorf("Failed to run state script: %v", err)
		}
	}

	util.RunCommandSequence(cfg.Bootcmd)

	if cfg.Rancher.State.Decrypt {
		cmdRun("ros", "console-init")
	}

	if cfg.Rancher.State.LvmScan {
		cmdRun("pvscan", "--activate", "ay")
		cmdRun("vgchange", "--activate", "ay")
	}

	if cfg.Rancher.State.Dev != "" && cfg.Rancher.State.Wait {
		waitForRoot(cfg)
	}

	autoformatDevices := cfg.Rancher.State.Autoformat
	if len(autoformatDevices) > 0 {
		if err := autoformat(autoformatDevices); err != nil {
			log.Errorf("Failed to run autoformat: %v", err)
		}
	}

	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}

	return nil
}

func cmdRun(cmdName string, args ...string) bool {
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Failed to run %v %v: %v", cmdName, args, err)
		return false
	}
	return true
}

func cmdStart(cmdName string, args ...string) error {
        cmd := exec.Command(cmdName, args...)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        return cmd.Start()
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
