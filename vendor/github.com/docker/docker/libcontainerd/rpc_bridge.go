package libcontainerd

import (
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/containerd/api/grpc/server"
	"github.com/docker/containerd/api/grpc/types"
	"github.com/docker/containerd/subreaper"
	"github.com/docker/containerd/supervisor"
)

var (
	stateDir = "/run/containerd"
)

type bridge struct {
	s types.APIServer
}

func newBridge(stateDir string, concurrency int, runtimeName string, runtimeArgs []string) (types.APIClient, error) {
	s, err := daemon(stateDir, concurrency, runtimeName, runtimeArgs)
	if err != nil {
		return nil, err
	}
	return &bridge{s: s}, nil
}

func daemon(stateDir string, concurrency int, runtimeName string, runtimeArgs []string) (types.APIServer, error) {
	if err := subreaper.Start(); err != nil {
		logrus.WithField("error", err).Error("containerd: start subreaper")
	}
	sv, err := supervisor.New(stateDir, runtimeName, "", runtimeArgs, 15*time.Second, 500)
	if err != nil {
		return nil, err
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		w := supervisor.NewWorker(sv, wg)
		go w.Start()
	}
	if err := sv.Start(); err != nil {
		return nil, err
	}
	return server.NewServer(sv), nil
}
