package power

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancherio/os/config"
)

func Main() {
	app := cli.NewApp()

	app.Name = os.Args[0]
	app.Usage = "Control and configure RancherOS"
	app.Version = config.VERSION
	app.Author = "Rancher Labs, Inc."
	app.EnableBashCompletion = true
	app.Action = shutdown
	app.HideHelp = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "r",
			Usage: "reboot after shutdown",
		},
		cli.BoolFlag{
			Name:  "H",
			Usage: "halt the system",
		},
		cli.BoolFlag{
			Name:  "P",
			Usage: "Power off the system",
		},
	}
	flags := []string{}
	for _, arg := range os.Args {
		if arg == "-h" {
			flags = append(flags, "-H")
		} else {
			flags = append(flags, arg)
		}
	}
	app.Run(flags)
}

func shutdown(c *cli.Context) {
	timeToParse := ""
	timeValue := c.Args().First()

	if len(timeValue) == 0 || timeValue == "" {
		log.Fatalf("No time value specified")
	}

	if timeValue[0] == '+' {
		if len(timeValue) > 1 {
			timeToParse = timeValue[1:]
		} else {
			log.Fatalf("invalid time value \"+\"")
		}
	} else if strings.Contains(timeValue, ":") {
		reference := "15:04"
		t, err := time.Parse(reference, timeValue)
		if err != nil {
			log.Fatalf("Could not parse time, err=%v", err)
		}
		now := time.Now()
		y, m, d := now.Date()
		a := t.AddDate(y, int(m)-1, d-1)
		b := a.Add(time.Duration(int(now.Day()-a.Day())*24) * time.Hour)
		mins := int(b.Sub(now).Minutes())
		if mins < 0 {
			mins = (24 * 60) + mins
		}
		timeToParse = strconv.Itoa(mins)
	} else if timeValue == "now" {
		timeToParse = "0"
	} else {
		log.Fatalf("invalid time value %s", timeValue)
	}
	timeout, err := strconv.Atoi(timeToParse)
	if err != nil {
		log.Fatalf("invalid time value %s", timeValue)
	}

	time.Sleep(time.Duration(timeout*60) * time.Second)
	common("")
	if os.ExpandEnv("${IN_DOCKER}") == "true" {
		rebootFlag := c.Bool("r")
		halt := c.Bool("H")
		_ = c.Bool("P")

		if rebootFlag {
			reboot(syscall.LINUX_REBOOT_CMD_RESTART, true, true)
		} else if halt {
			reboot(syscall.LINUX_REBOOT_CMD_HALT, true, true)
		} else {
			reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF, true, true)
		}
	}
}
