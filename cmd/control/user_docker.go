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

	"github.com/codegangsta/cli"
	//composeClient "github.com/docker/libcompose/docker/client"
	"github.com/docker/libcompose/project"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	//rosDocker "github.com/rancher/os/docker"
	"github.com/rancher/os/log"
	"github.com/rancher/os/util"

	"fmt"
	//	"github.com/containerd/console"
	"github.com/containerd/containerd"
	tasks "github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types/task"
	"github.com/containerd/containerd/namespaces"
)

const (
	defaultStorageContext = "console"
	dockerPidFile         = "/var/run/docker.pid"
	sourceDirectory       = "/engine"
	destDirectory         = "/var/lib/rancher/engine"
)

var (
	dockerCommand = []string{
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
	outDir := ServerTLSPath
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
		storageContext = defaultStorageContext
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

	dockerCfg := cfg.Rancher.Docker

	args := dockerCfg.FullArgs()

	log.Debugf("User Docker args: %v", args)

	if dockerCfg.TLS {
		if err := writeConfigCerts(cfg); err != nil {
			return err
		}
	}

	client, err := containerd.New(config.DefaultContainerdSocket)
	if err != nil {
		log.Errorf("creating containerd client: %s", err)
	}
	ctx := namespaces.WithNamespace(context.Background(), "default")
	container, err := client.LoadContainer(ctx, storageContext)
	if err != nil {
		return err
	}
	spec, err := container.Spec()
	if err != nil {
		return err
	}
	task, err := container.Task(ctx, nil)
	if err != nil {
		return err
	}

	pspec := spec.Process
	//pspec.Terminal = tty
	pspec.Args = []string{}
	//cmd := []string{"docker-runc", "exec", "--", info.ID, "env"}
	//	log.Info(dockerCfg.AppendEnv())
	//	pspec.Args = append(pspec.Args, dockerCfg.AppendEnv()...)
	pspec.Args = append(pspec.Args, dockerCommand...)
	pspec.Args = append(pspec.Args, args...)
	log.Infof("Running %v", pspec.Args)

	io := containerd.Stdio
	tty := false
	if tty {
		io = containerd.StdioTerminal
	}
	process, err := task.Exec(ctx, "docker-exec", pspec, io)
	if err != nil {
		log.Infof("Error creating process: %s", err)

		return err
	}
	defer process.Delete(ctx)

	statusC, err := process.Wait(ctx)
	if err != nil {
		log.Infof("Error waiting: %s", err)
		return err
	}
	log.Infof("STARTED(%s): %s\n", pspec.Args, statusC)

	//var con console.Console
	//if tty {
	//	con = console.Current()
	//	defer con.Reset()
	//	if err := con.SetRaw(); err != nil {
	//		return err
	//	}
	//}
	//if tty {
	//if err := handleConsoleResize(ctx, process, con); err != nil {
	//	log.WithError("console resize: %s", err)
	//}
	//} else {
	//	sigc := forwardAllSignals(ctx, process)
	//	defer stopCatch(sigc)
	//}

	if err := process.Start(ctx); err != nil {
		log.Infof("Error starting process: %s", err)
		return err
	}
	status := <-statusC
	code, _, err := status.Result()
	fmt.Printf("FINISHED (%s): %s\n", pspec.Args, statusC)
	if err != nil {
		return err
	}
	if code != 0 {
		return cli.NewExitError("", int(code))
	}

	//cmd := []string{"docker-runc", "exec", "--", info.ID, "env"}
	//log.Info(dockerCfg.AppendEnv())
	//cmd = append(cmd, dockerCfg.AppendEnv()...)
	//cmd = append(cmd, dockerCommand...)
	//cmd = append(cmd, args...)
	//log.Infof("Running %v", cmd)

	//return syscall.Exec("/usr/bin/ros", cmd, os.Environ())
	return nil
}

func waitForPid(service string, project *project.Project) (uint32, error) {
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

func getPid(service string, project *project.Project) (uint32, error) {
	client, err := containerd.New(config.DefaultContainerdSocket)
	if err != nil {
		log.Errorf("creating containerd client: %s", err)
	}
	ctx := namespaces.WithNamespace(context.Background(), "default")

	s := client.TaskService()
	response, err := s.List(ctx, &tasks.ListTasksRequest{})
	if err != nil {
		return 0, err
	}
	for _, t := range response.Tasks {
		if t.ID == service && t.Status == task.StatusRunning {
			return t.Pid, nil
		}
	}

	return 0, fmt.Errorf("service task (%s) not running", service)
}
