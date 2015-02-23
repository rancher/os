package power

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsouza/go-dockerclient"
	log "github.com/Sirupsen/logrus"
)

const (
	dockerPath        = "unix:///var/run/system-docker.sock"
	dockerCGroupsFile = "/proc/self/cgroup"
)

func PowerOff() {
	defer func() {
		if r := recover(); r!= nil {
			log.Info("Error while shutting down containers, gracefully stopping")
			os.Exit(0)
		}
	}()
	if os.Geteuid() != 0 {
		log.Info("poweroff: Permission Denied")
		return	
	}
	syscall.Sync()
	reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func Reboot() {
	defer func() {
		if r := recover(); r!= nil {
			log.Info("Error while shutting down containers, gracefully stopping")
			os.Exit(0)
		}
	}()
	if os.Geteuid() != 0 {
		log.Info("reboot: Permission Denied")
		return
	}
	syscall.Sync()
	reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func Halt() {
	defer func() {
		if r := recover(); r!= nil {
			log.Info("Error while shutting down containers, gracefully stopping")
			os.Exit(0)
		}
	}()
	if os.Geteuid() != 0 {
		log.Info("reboot: Permission Denied")
		return
	}
	syscall.Sync()
	reboot(syscall.LINUX_REBOOT_CMD_HALT)
}

func reboot(code int) {
	err := shutDownContainers()
	if err != nil {
		log.Fatal(err)
	}
	err = syscall.Reboot(code)
	if err != nil {
		log.Fatal(err)
	}
}

func shutDownContainers() error {
	var err error
	shutDown := true
	timeout := uint(2)
	for i := range os.Args {
		arg := os.Args[i]
		if arg == "-f" || arg == "--f" || arg == "--force" {
			shutDown = false
		}
		if arg == "-t" || arg == "--t" || arg == "--timeout" {
			if len(os.Args) > i+1 {
				t, er := strconv.Atoi(os.Args[i+1])
				if er != nil {
					return err
				}
				timeout = uint(t)
			} else {
				log.Info("please specify a timeout")
			}
		}
	}
	if !shutDown {
		return nil
	}
	client, err := docker.NewClient(dockerPath)

	if err != nil {
		return err
	}

	opts := docker.ListContainersOptions{All: true, Filters: map[string][]string{"status": []string{"running"}}}
	var containers []docker.APIContainers

	containers, err = client.ListContainers(opts)

	if err != nil {
		return err
	}

	currentContainerId, err := getCurrentContainerId()

	if err != nil {
		return err
	}

	var stopErrorStrings []string

	for i := range containers {
		if containers[i].ID == currentContainerId {
			continue
		}
		stopErr := client.StopContainer(containers[i].ID, timeout)
		if stopErr != nil {
			stopErrorStrings = append(stopErrorStrings, " ["+containers[i].ID+"] "+stopErr.Error())
		}
	}

	var waitErrorStrings []string

	for i := range containers {
		if containers[i].ID == currentContainerId {
			continue
		}
		_, waitErr := client.WaitContainer(containers[i].ID)
		if waitErr != nil {
			waitErrorStrings = append(waitErrorStrings, " ["+containers[i].ID+"] "+waitErr.Error())
		}
	}

	if len(waitErrorStrings) != 0 || len(stopErrorStrings) != 0 {
		return errors.New("error while stopping \n1. STOP Errors [" + strings.Join(stopErrorStrings, ",") + "] \n2. WAIT Errors [" + strings.Join(waitErrorStrings, ",") + "]")
	}

	return nil
}

func getCurrentContainerId() (string, error) {
	file, err := os.Open(dockerCGroupsFile)

	if err != nil {
		return "", err
	}

	fileReader := bufio.NewScanner(file)
	if !fileReader.Scan() {
		return "", errors.New("Empty file /proc/self/cgroup")
	}
	line := fileReader.Text()
	parts := strings.Split(line, "/")

	for len(parts) != 3 {
		if !fileReader.Scan() {
			return "", errors.New("Found no docker cgroups")
		}
		line = fileReader.Text()
		parts = strings.Split(line, "/")
		if len(parts) == 3 {
			if strings.HasSuffix(parts[1], "docker") {
				break
			} else {
				parts = nil
			}
		}
	}

	return parts[len(parts)-1:][0], nil
}
