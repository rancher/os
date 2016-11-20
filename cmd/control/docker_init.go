package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util"
)

const (
	dockerConf = "/var/lib/rancher/conf/docker"
	dockerDone = "/run/docker-done"
	dockerLog  = "/var/log/docker.log"
)

func dockerInitAction(c *cli.Context) error {
	for {
		if _, err := os.Stat(consoleDone); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	dockerBin := "/usr/bin/docker"
	for _, binPath := range []string{
		"/opt/bin",
		"/usr/local/bin",
		"/var/lib/rancher/docker",
	} {
		if util.ExistsAndExecutable(path.Join(binPath, "dockerd")) {
			dockerBin = path.Join(binPath, "dockerd")
			break
		}
		if util.ExistsAndExecutable(path.Join(binPath, "docker")) {
			dockerBin = path.Join(binPath, "docker")
			break
		}
	}

	if err := syscall.Mount("", "/", "", syscall.MS_SHARED|syscall.MS_REC, ""); err != nil {
		log.Error(err)
	}
	if err := syscall.Mount("", "/run", "", syscall.MS_SHARED|syscall.MS_REC, ""); err != nil {
		log.Error(err)
	}

	mountInfo, err := ioutil.ReadFile("/proc/self/mountinfo")
	if err != nil {
		return err
	}

	for _, mount := range strings.Split(string(mountInfo), "\n") {
		if strings.Contains(mount, "/var/lib/docker /var/lib/docker") && strings.Contains(mount, "rootfs") {
			os.Setenv("DOCKER_RAMDISK", "1")
		}
	}

	args := []string{
		"bash",
		"-c",
		fmt.Sprintf(`[ -e %s ] && source %s; exec /usr/bin/dockerlaunch %s %s $DOCKER_OPTS >> %s 2>&1`, dockerConf, dockerConf, dockerBin, strings.Join(c.Args(), " "), dockerLog),
	}

	cfg := config.LoadConfig()

	if err := ioutil.WriteFile(dockerDone, []byte(cfg.Rancher.Docker.Engine), 0644); err != nil {
		log.Error(err)
	}

	return syscall.Exec("/bin/bash", args, os.Environ())
}
