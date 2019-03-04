package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"

	"github.com/codegangsta/cli"
)

const (
	dockerConf               = "/var/lib/rancher/conf/docker"
	dockerDone               = "/run/docker-done"
	dockerLog                = "/var/log/docker.log"
	dockerCompletionLinkFile = "/usr/share/bash-completion/completions/docker"
	dockerCompletionFile     = "/var/lib/rancher/engine/completion"
)

func dockerInitAction(c *cli.Context) error {
	// TODO: this should be replaced by a "Console ready event watcher"
	for {
		if _, err := os.Stat(consoleDone); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if _, err := os.Stat(dockerCompletionFile); err != nil {
		if _, err := os.Readlink(dockerCompletionLinkFile); err == nil {
			syscall.Unlink(dockerCompletionLinkFile)
		}
	}

	dockerBin := ""
	dockerPaths := []string{
		"/usr/bin",
		"/opt/bin",
		"/usr/local/bin",
		"/var/lib/rancher/docker",
	}
	for _, binPath := range dockerPaths {
		if util.ExistsAndExecutable(path.Join(binPath, "dockerd")) {
			dockerBin = path.Join(binPath, "dockerd")
			break
		}
	}
	if dockerBin == "" {
		for _, binPath := range dockerPaths {
			if util.ExistsAndExecutable(path.Join(binPath, "docker")) {
				dockerBin = path.Join(binPath, "docker")
				break
			}
		}
	}
	if dockerBin == "" {
		err := fmt.Errorf("Failed to find either dockerd or docker binaries")
		log.Error(err)
		return err
	}
	log.Infof("Found %s", dockerBin)

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
		if strings.Contains(mount, "/var/lib/user-docker /var/lib/docker") && strings.Contains(mount, "rootfs") {
			os.Setenv("DOCKER_RAMDISK", "true")
		}
	}

	cfg := config.LoadConfig()
	baseSymlink := symLinkEngineBinary(cfg.Rancher.Docker.Engine)

	for _, link := range baseSymlink {
		syscall.Unlink(link.newname)
		if err := os.Symlink(link.oldname, link.newname); err != nil {
			log.Error(err)
		}
	}

	err = checkZfsBackingFS(cfg.Rancher.Docker.StorageDriver, cfg.Rancher.Docker.Graph)
	if err != nil {
		log.Fatal(err)
	}

	args := []string{
		"bash",
		"-c",
		fmt.Sprintf(`[ -e %s ] && source %s; exec /usr/bin/dockerlaunch %s %s $DOCKER_OPTS >> %s 2>&1`, dockerConf, dockerConf, dockerBin, strings.Join(c.Args(), " "), dockerLog),
	}

	// TODO: this should be replaced by a "Docker ready event watcher"
	if err := ioutil.WriteFile(dockerDone, []byte(CurrentEngine()), 0644); err != nil {
		log.Error(err)
	}

	return syscall.Exec("/bin/bash", args, os.Environ())
}
