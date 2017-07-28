package power

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/cmd/control/install"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

var (
	haltFlag          bool
	poweroffFlag      bool
	rebootFlag        bool
	forceFlag         bool
	kexecFlag         bool
	previouskexecFlag bool
	kexecAppendFlag   string
)

func Shutdown() {
	log.InitLogger()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = fmt.Sprintf("%s RancherOS\nbuilt: %s", app.Name, config.BuildDate)
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Action = shutdown
	app.Flags = []cli.Flag{
		//    --no-wall
		//        Do not send wall message before halt, power-off,
		//        reboot.

		// halt, poweroff, reboot ONLY
		//    -f, --force
		//        Force immediate halt, power-off, reboot. Do not
		//        contact the init system.
		cli.BoolFlag{
			Name:        "f, force",
			Usage:       "Force immediate halt, power-off, reboot. Do not contact the init system.",
			Destination: &forceFlag,
		},

		//    -w, --wtmp-only
		//        Only write wtmp shutdown entry, do not actually
		//        halt, power-off, reboot.

		//    -d, --no-wtmp
		//        Do not write wtmp shutdown entry.

		//    -n, --no-sync
		//        Don't sync hard disks/storage media before halt,
		//        power-off, reboot.

		// shutdown ONLY
		//    -h
		//        Equivalent to --poweroff, unless --halt is
		//        specified.

		//    -k
		//        Do not halt, power-off, reboot, just write wall
		//        message.

		//    -c
		//        Cancel a pending shutdown. This may be used
		//        cancel the effect of an invocation of shutdown
		//        with a time argument that is not "+0" or "now".

	}
	//    -H, --halt
	//        Halt the machine.
	if app.Name == "halt" {
		app.Flags = append(app.Flags, cli.BoolTFlag{
			Name:        "H, halt",
			Usage:       "halt the machine",
			Destination: &haltFlag,
		})
	} else {
		app.Flags = append(app.Flags, cli.BoolFlag{
			Name:        "H, halt",
			Usage:       "halt the machine",
			Destination: &haltFlag,
		})
	}
	//    -P, --poweroff
	//        Power-off the machine (the default for shutdown cmd).
	if app.Name == "poweroff" {
		app.Flags = append(app.Flags, cli.BoolTFlag{
			Name:        "P, poweroff",
			Usage:       "halt the machine",
			Destination: &poweroffFlag,
		})
	} else {
		app.Flags = append(app.Flags, cli.BoolFlag{
			Name:        "P, poweroff",
			Usage:       "halt the machine",
			Destination: &poweroffFlag,
		})
	}
	//    -r, --reboot
	//        Reboot the machine.
	if app.Name == "reboot" {
		app.Flags = append(app.Flags, cli.BoolTFlag{
			Name:        "r, reboot",
			Usage:       "reboot after shutdown",
			Destination: &rebootFlag,
		})
		// OR? maybe implement it as a `kexec` cli tool?
		app.Flags = append(app.Flags, cli.BoolFlag{
			Name:        "kexec",
			Usage:       "kexec the default RancherOS cfg",
			Destination: &kexecFlag,
		})
		app.Flags = append(app.Flags, cli.BoolFlag{
			Name:        "kexec-previous",
			Usage:       "kexec the previous RancherOS cfg",
			Destination: &previouskexecFlag,
		})
		app.Flags = append(app.Flags, cli.StringFlag{
			Name:        "kexec-append",
			Usage:       "kexec using the specified kernel boot params (ignores global.cfg)",
			Destination: &kexecAppendFlag,
		})
	} else {
		app.Flags = append(app.Flags, cli.BoolFlag{
			Name:        "r, reboot",
			Usage:       "reboot after shutdown",
			Destination: &rebootFlag,
		})
	}
	//TODO: add the time and msg flags...
	app.HideHelp = true

	app.Run(os.Args)
}

func Kexec(previous bool, bootDir, append string) error {
	cfg := "linux-current.cfg"
	if previous {
		cfg = "linux-previous.cfg"
	}
	cfgFile := filepath.Join(bootDir, cfg)
	vmlinuzFile, initrdFile, err := install.ReadSyslinuxCfg(cfgFile)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}
	globalCfgFile := filepath.Join(bootDir, "global.cfg")
	if append == "" {
		append, err = install.ReadGlobalCfg(globalCfgFile)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
	}
	// TODO: read global.cfg if append == ""
	//    kexec -l ${DIST}/vmlinuz --initrd=${DIST}/initrd --append="${kernelArgs} ${APPEND}" -f
	cmd := exec.Command(
		"kexec",
		"-l", vmlinuzFile,
		"--initrd", initrdFile,
		"--append", append,
		"-f")
	log.Debugf("Run(%#v)", cmd)
	cmd.Stderr = os.Stderr
	if _, err := cmd.Output(); err != nil {
		log.Errorf("Failed to kexec: %s", err)
		return err
	}
	log.Infof("kexec'd to new install")
	return nil
}

// Reboot is used by installation / upgrade
// TODO: add kexec option
func Reboot() {
	reboot("reboot", false, syscall.LINUX_REBOOT_CMD_RESTART)
}

func shutdown(c *cli.Context) error {
	// the shutdown command's default is poweroff
	var powerCmd uint
	powerCmd = syscall.LINUX_REBOOT_CMD_POWER_OFF
	if rebootFlag {
		powerCmd = syscall.LINUX_REBOOT_CMD_RESTART
	} else if poweroffFlag {
		powerCmd = syscall.LINUX_REBOOT_CMD_POWER_OFF
	} else if haltFlag {
		powerCmd = syscall.LINUX_REBOOT_CMD_HALT
	}

	timeArg := c.Args().Get(0)
	if c.App.Name == "shutdown" && timeArg != "" {
		if timeArg != "now" {
			err := fmt.Errorf("Sorry, can't parse '%s' as time value (only 'now' supported)", timeArg)
			log.Error(err)
			return err
		}
		// TODO: if there are more params, LOG them
	}

	reboot(c.App.Name, forceFlag, powerCmd)

	return nil
}
