package control

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	composeClient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	rosDocker "github.com/rancher/os/docker"
	"github.com/rancher/os/util"
)

const (
	DEFAULT_STORAGE_CONTEXT = "console"
	DOCKER_PID_FILE         = "/var/run/docker.pid"
	userDocker              = "user-docker"
	sourceDirectory         = "/engine"
	destDirectory           = "/var/lib/rancher/engine"
)

var (
	DOCKER_COMMAND = []string{
		"ros",
		"docker-init",
	}
)

func userDockerAction(c *cli.Context) error {
	if err := copyBinaries(sourceDirectory, destDirectory); err != nil {
		return err
	}

	if err := syscall.Mount("/host/sys", "/sys", "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return err
	}

	cfg := config.LoadConfig()

	return startDocker(cfg)
}

func copyBinaries(source, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(dest)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err = os.RemoveAll(path.Join(dest, file.Name())); err != nil {
			return err
		}
	}

	files, err = ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	for _, file := range files {
		sourceFile := path.Join(source, file.Name())
		destFile := path.Join(dest, file.Name())

		in, err := os.Open(sourceFile)
		if err != nil {
			return err
		}
		out, err := os.Create(destFile)
		if err != nil {
			return err
		}
		if _, err = io.Copy(out, in); err != nil {
			return err
		}
		if err = out.Sync(); err != nil {
			return err
		}
		if err = in.Close(); err != nil {
			return err
		}
		if err = out.Close(); err != nil {
			return err
		}
		if err := os.Chmod(destFile, 0751); err != nil {
			return err
		}
	}

	return nil
}

func writeConfigCerts(cfg *config.CloudConfig) error {
	outDir := ServerTlsPath
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return err
	}
	caCertPath := filepath.Join(outDir, CaCert)
	caKeyPath := filepath.Join(outDir, CaKey)
	serverCertPath := filepath.Join(outDir, ServerCert)
	serverKeyPath := filepath.Join(outDir, ServerKey)
	if cfg.Rancher.Docker.CACert != "" {
		if err := util.WriteFileAtomic(caCertPath, []byte(cfg.Rancher.Docker.CACert), 0400); err != nil {
			return err
		}

		if err := util.WriteFileAtomic(caKeyPath, []byte(cfg.Rancher.Docker.CAKey), 0400); err != nil {
			return err
		}
	}
	if cfg.Rancher.Docker.ServerCert != "" {
		if err := util.WriteFileAtomic(serverCertPath, []byte(cfg.Rancher.Docker.ServerCert), 0400); err != nil {
			return err
		}

		if err := util.WriteFileAtomic(serverKeyPath, []byte(cfg.Rancher.Docker.ServerKey), 0400); err != nil {
			return err
		}
	}
	return nil
}

func startDocker(cfg *config.CloudConfig) error {
	storageContext := cfg.Rancher.Docker.StorageContext
	if storageContext == "" {
		storageContext = DEFAULT_STORAGE_CONTEXT
	}

	log.Infof("Starting Docker in context: %s", storageContext)

	p, err := compose.GetProject(cfg, true, false)
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

	dockerCfg := cfg.Rancher.Docker

	args := dockerCfg.FullArgs()

	log.Debugf("User Docker args: %v", args)

	if dockerCfg.TLS {
		if err := writeConfigCerts(cfg); err != nil {
			return err
		}
	}

	info, err := client.ContainerInspect(context.Background(), storageContext)
	if err != nil {
		return err
	}

	cmd := []string{"docker-runc", "exec", "--", info.ID, "env"}
	log.Info(dockerCfg.AppendEnv())
	cmd = append(cmd, dockerCfg.AppendEnv()...)
	cmd = append(cmd, DOCKER_COMMAND...)
	cmd = append(cmd, args...)
	log.Infof("Running %v", cmd)

	return syscall.Exec("/usr/bin/ros", cmd, os.Environ())
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
