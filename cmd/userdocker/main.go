package userdocker

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	composeClient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/cmd/control"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	rosDocker "github.com/rancher/os/docker"
)

const (
	DEFAULT_STORAGE_CONTEXT = "console"
	DOCKER_PID_FILE         = "/var/run/docker.pid"
	DOCKER_COMMAND          = "docker-init"
	userDocker              = "user-docker"
)

func Main() {
	cfg := config.LoadConfig()

	if err := startDocker(cfg); err != nil {
		log.Fatal(err)
	}

	if err := setupTermHandler(); err != nil {
		log.Fatal(err)
	}

	select {}
}

func startDocker(cfg *config.CloudConfig) error {
	storageContext := cfg.Rancher.Docker.StorageContext
	if storageContext == "" {
		storageContext = DEFAULT_STORAGE_CONTEXT
	}

	log.Infof("Starting Docker in context: %s", storageContext)

	p, err := compose.GetProject(cfg, true)
	if err != nil {
		return err
	}

	pid, err := waitForPid(storageContext, p)
	if err != nil {
		return err
	}

	log.Infof("%s PID %d", storageContext, pid)

	client, err := rosDocker.NewSystemClient()
	if err != nil {
		return err
	}

	if err := os.Remove(DOCKER_PID_FILE); err != nil && !os.IsNotExist(err) {
		return err
	}

	dockerCfg := cfg.Rancher.Docker

	args := dockerCfg.FullArgs()

	log.Debugf("User Docker args: %v", args)

	if dockerCfg.TLS {
		log.Debug("Generating TLS certs if needed")
		if err := control.Generate(true, "/etc/docker/tls", []string{"127.0.0.1", "*", "*.*", "*.*.*", "*.*.*.*"}); err != nil {
			return err
		}
	}

	cmd := []string{"env"}
	log.Info(dockerCfg.AppendEnv())
	cmd = append(cmd, dockerCfg.AppendEnv()...)
	cmd = append(cmd, DOCKER_COMMAND)
	cmd = append(cmd, args...)
	log.Infof("Running %v", cmd)

	resp, err := client.ContainerExecCreate(context.Background(), types.ExecConfig{
		Container:    storageContext,
		Privileged:   true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Detach:       false,
		Cmd:          cmd,
	})
	if err != nil {
		return err
	}

	if err := client.ContainerExecStart(context.Background(), resp.ID, types.ExecStartCheck{
		Detach: false,
	}); err != nil {
		return err
	}

	return nil
}

func setupTermHandler() error {
	pidBytes, err := waitForFile(DOCKER_PID_FILE)
	if err != nil {
		return err
	}
	dockerPid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return err
	}
	process, err := os.FindProcess(dockerPid)
	if err != nil {
		return err
	}
	handleTerm(process)
	return nil
}

func handleTerm(p *os.Process) {
	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGTERM)
	go func() {
		<-term
		p.Signal(syscall.SIGTERM)
		os.Exit(0)
	}()
}

func waitForFile(file string) ([]byte, error) {
	for {
		contents, err := ioutil.ReadFile(file)
		if os.IsNotExist(err) {
			log.Infof("Waiting for %s", file)
			time.Sleep(1 * time.Second)
		} else if err != nil {
			return nil, err
		} else {
			return contents, nil
		}
	}
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

	containers, err := s.Containers(context.Background())
	if err != nil {
		return 0, err
	}

	if len(containers) == 0 {
		return 0, nil
	}

	client, err := composeClient.Create(composeClient.Options{
		Host: config.DOCKER_SYSTEM_HOST,
	})
	if err != nil {
		return 0, err
	}

	id, err := containers[0].ID()
	if err != nil {
		return 0, err
	}

	info, err := client.ContainerInspect(context.Background(), id)
	if err != nil || info.ID == "" {
		return 0, err
	}

	if info.State.Running {
		return info.State.Pid, nil
	}

	return 0, nil
}
