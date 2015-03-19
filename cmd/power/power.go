package power

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	dockerClient "github.com/fsouza/go-dockerclient"

	"github.com/rancherio/os/docker"
)

const (
	DOCKER_CGROUPS_FILE = "/proc/self/cgroup"
)

func runDocker(name string) error {
	if os.ExpandEnv("${IN_DOCKER}") == "true" {
		return nil
	}

	client, err := docker.NewSystemClient()
	if err != nil {
		return err
	}

	cmd := []string{name}

	if name == "" {
		name = filepath.Base(os.Args[0])
		cmd = os.Args
	}

	exiting, err := client.InspectContainer(name)
	if exiting != nil {
		err := client.RemoveContainer(dockerClient.RemoveContainerOptions{
			ID:    exiting.ID,
			Force: true,
		})

		if err != nil {
			return err
		}
	}

	currentContainerId, err := getCurrentContainerId()
	if err != nil {
		return err
	}

	currentContainer, err := client.InspectContainer(currentContainerId)
	if err != nil {
		return err
	}

	powerContainer, err := client.CreateContainer(dockerClient.CreateContainerOptions{
		Name: name,
		Config: &dockerClient.Config{
			Image: currentContainer.Config.Image,
			Cmd:   cmd,
			Env: []string{
				"IN_DOCKER=true",
			},
		},
		HostConfig: &dockerClient.HostConfig{
			PidMode: "host",
			VolumesFrom: []string{
				currentContainer.ID,
			},
			Privileged: true,
		},
	})
	if err != nil {
		return err
	}

	go func() {
		client.AttachToContainer(dockerClient.AttachToContainerOptions{
			Container:    powerContainer.ID,
			OutputStream: os.Stdout,
			ErrorStream:  os.Stderr,
			Stderr:       true,
			Stdout:       true,
		})
	}()

	err = client.StartContainer(powerContainer.ID, powerContainer.HostConfig)
	if err != nil {
		return err
	}

	_, err = client.WaitContainer(powerContainer.ID)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
	return nil
}

func common(name string) {
	if os.Geteuid() != 0 {
		log.Fatalf("%s: Need to be root", os.Args[0])
	}

	if err := runDocker(name); err != nil {
		log.Fatal(err)
	}
}

func PowerOff() {
	common("poweroff")
	reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func Reboot() {
	common("reboot")
	reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func Halt() {
	common("halt")
	reboot(syscall.LINUX_REBOOT_CMD_HALT)
}

func reboot(code int) {
	err := shutDownContainers()
	if err != nil {
		log.Error(err)
	}

	syscall.Sync()

	err = syscall.Reboot(code)
	if err != nil {
		log.Fatal(err)
	}
}

func shutDownContainers() error {
	var err error
	shutDown := true
	timeout := 2
	for i, arg := range os.Args {
		if arg == "-f" || arg == "--f" || arg == "--force" {
			shutDown = false
		}
		if arg == "-t" || arg == "--t" || arg == "--timeout" {
			if len(os.Args) > i+1 {
				t, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					return err
				}
				timeout = t
			} else {
				log.Error("please specify a timeout")
			}
		}
	}
	if !shutDown {
		return nil
	}
	client, err := docker.NewSystemClient()

	if err != nil {
		return err
	}

	opts := dockerClient.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"status": []string{"running"},
		},
	}

	containers, err := client.ListContainers(opts)
	if err != nil {
		return err
	}

	currentContainerId, err := getCurrentContainerId()
	if err != nil {
		return err
	}

	var stopErrorStrings []string

	for _, container := range containers {
		if container.ID == currentContainerId {
			continue
		}

		log.Infof("Stopping %s : %v", container.ID[:12], container.Names)
		stopErr := client.StopContainer(container.ID, uint(timeout))
		if stopErr != nil {
			stopErrorStrings = append(stopErrorStrings, " ["+container.ID+"] "+stopErr.Error())
		}
	}

	var waitErrorStrings []string

	for _, container := range containers {
		if container.ID == currentContainerId {
			continue
		}
		_, waitErr := client.WaitContainer(container.ID)
		if waitErr != nil {
			waitErrorStrings = append(waitErrorStrings, " ["+container.ID+"] "+waitErr.Error())
		}
	}

	if len(waitErrorStrings) != 0 || len(stopErrorStrings) != 0 {
		return errors.New("error while stopping \n1. STOP Errors [" + strings.Join(stopErrorStrings, ",") + "] \n2. WAIT Errors [" + strings.Join(waitErrorStrings, ",") + "]")
	}

	return nil
}

func getCurrentContainerId() (string, error) {
	file, err := os.Open(DOCKER_CGROUPS_FILE)

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
