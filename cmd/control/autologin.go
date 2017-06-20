package control

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
	"github.com/rancher/os/log"
)

func AutologinMain() {
	log.InitLogger()
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "autologin console"
	app.Version = config.Version
	app.Author = "Rancher Labs, Inc."
	app.Email = "sven@rancher.com"
	app.EnableBashCompletion = true
	app.Action = autologinAction
	app.HideHelp = true
	app.Run(os.Args)
}


func autologinAction(c *cli.Context) error {
	cmd := exec.Command("/bin/stty", "sane")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}

	usertty := ""
	user := "root"
	tty := ""
	if c.NArg() > 0 {
		usertty = c.Args().Get(0)
		s := strings.SplitN(usertty, ":", 2)
		user = s[0]
		if len(s) > 1 {
			tty = s[1]
		}
	}
	cfg := config.LoadConfig()
	// replace \n and \l
	banner := config.Banner
	banner = strings.Replace(banner, "\\v", config.Version, -1)
	banner = strings.Replace(banner, "\\s", "RancherOS " + runtime.GOARCH, -1)
	banner = strings.Replace(banner, "\\r", "4.9....", -1)
	banner = strings.Replace(banner, "\\n", cfg.Hostname, -1)
	banner = strings.Replace(banner, "\\l", tty, -1)
	banner = strings.Replace(banner, "\\\\", "\\", -1)
	banner = banner + "\n"
	fmt.Printf(banner)

	cmd = exec.Command("/usr/bin/login", "-f", user)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Error(err)
	}
	return nil
}

