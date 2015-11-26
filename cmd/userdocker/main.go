package userdocker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"

	"github.com/opencontainers/runc/libcontainer/cgroups"
	_ "github.com/opencontainers/runc/libcontainer/nsenter"
	"github.com/opencontainers/runc/libcontainer/system"
)

const (
	DEFAULT_STORAGE_CONTEXT = "console"
	userDocker              = "user-docker"
)

func Main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 1 {
		if err := enter(cfg); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := main(cfg); err != nil {
			log.Fatal(err)
		}
	}
}

func enter(cfg *config.CloudConfig) error {
	context := cfg.Rancher.Docker.StorageContext
	if context == "" {
		context = DEFAULT_STORAGE_CONTEXT
	}

	log.Infof("Starting Docker in context: %s", context)

	p, err := compose.GetProject(cfg, true)
	if err != nil {
		return err
	}

	pid, err := waitForPid(context, p)
	if err != nil {
		return err
	}

	log.Infof("%s PID %d", context, pid)

	return runNsenter(pid)
}

type result struct {
	Pid int `json:"Pid"`
}

func findProgram(searchPaths ...string) string {
	prog := ""

	for _, i := range searchPaths {
		var err error
		prog, err = exec.LookPath(i)
		if err == nil {
			break
		}
		prog = i
	}

	return prog
}

func runNsenter(pid int) error {
	args := []string{findProgram(userDocker), "main"}

	r, w, err := os.Pipe()
	if err != nil {
		return err
	}

	cmd := &exec.Cmd{
		Path:       args[0],
		Args:       args,
		Stdin:      os.Stdin,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
		ExtraFiles: []*os.File{w},
		Env: append(os.Environ(),
			"_LIBCONTAINER_INITPIPE=3",
			fmt.Sprintf("_LIBCONTAINER_INITPID=%d", pid),
		),
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	w.Close()

	var result result
	if err := json.NewDecoder(r).Decode(&result); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	log.Infof("Docker PID %d", result.Pid)

	p, err := os.FindProcess(result.Pid)
	if err != nil {
		return err
	}

	handleTerm(p)

	if err := switchCgroup(result.Pid, pid); err != nil {
		return err
	}

	_, err = p.Wait()
	return err
}

func handleTerm(p *os.Process) {
	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM)
	go func() {
		<-term
		p.Signal(syscall.SIGTERM)
	}()
}

func waitForPid(service string, project *project.Project) (int, error) {
	log.Infof("Getting PID for service: %s", service)
	for {
		if pid, err := getPid(service, project); err != nil || pid == 0 {
			log.Infof("Waiting for %s : %d : %v", service, pid, err)
			time.Sleep(1 * time.Second)
		} else {
			return pid, err
		}
	}
}

func getPid(service string, project *project.Project) (int, error) {
	s, err := project.CreateService(service)
	if err != nil {
		return 0, err
	}

	containers, err := s.Containers()
	if err != nil {
		return 0, err
	}

	if len(containers) == 0 {
		return 0, nil
	}

	client, err := docker.CreateClient(docker.ClientOpts{
		Host: config.DOCKER_SYSTEM_HOST,
	})
	if err != nil {
		return 0, err
	}

	id, err := containers[0].ID()
	if err != nil {
		return 0, err
	}

	info, err := client.InspectContainer(id)
	if err != nil || info == nil {
		return 0, err
	}

	if info.State.Running {
		return info.State.Pid, nil
	}

	return 0, nil
}

func main(cfg *config.CloudConfig) error {
	os.Unsetenv("_LIBCONTAINER_INITPIPE")
	os.Unsetenv("_LIBCONTAINER_INITPID")

	if err := system.ParentDeathSignal(syscall.SIGKILL).Set(); err != nil {
		return err
	}

	if err := os.Remove("/var/run/docker.pid"); err != nil && !os.IsNotExist(err) {
		return err
	}

	dockerCfg := cfg.Rancher.Docker

	args := dockerCfg.FullArgs()

	log.Debugf("User Docker args: %v", args)

	if dockerCfg.TLS {
		log.Debug("Generating TLS certs if needed")
		if err := control.Generate(true, "/etc/docker/tls", []string{"localhost"}); err != nil {
			return err
		}
	}

	prog := findProgram("docker-init", "dockerlaunch", "docker")
	if strings.Contains(prog, "dockerlaunch") {
		args = append([]string{prog, "docker"}, args...)
	} else {
		args = append([]string{prog}, args...)
	}

	log.Infof("Running %v", args)
	return syscall.Exec(args[0], args, dockerCfg.AppendEnv())
}

func switchCgroup(src, target int) error {
	cgroupFile := fmt.Sprintf("/proc/%d/cgroup", target)
	f, err := os.Open(cgroupFile)
	if err != nil {
		return err
	}
	defer f.Close()

	targetCgroups := map[string]string{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		parts := strings.Split(text, ":")
		subparts := strings.Split(parts[1], "=")
		subsystem := subparts[0]
		if len(subparts) > 1 {
			subsystem = subparts[1]
		}

		targetPath := fmt.Sprintf("/host/sys/fs/cgroup/%s%s", subsystem, parts[2])
		log.Infof("Moving Docker to cgroup %s", targetPath)
		targetCgroups[subsystem] = targetPath
	}

	if err := s.Err(); err != nil {
		return err
	}

	return cgroups.EnterPid(targetCgroups, src)
}
