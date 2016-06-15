package libcontainerd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	containerd "github.com/docker/containerd/api/grpc/types"
	"github.com/docker/docker/pkg/locker"
	sysinfo "github.com/docker/docker/pkg/system"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/transport"
)

const (
	maxConnectionRetryCount   = 3
	connectionRetryDelay      = 3 * time.Second
	containerdShutdownTimeout = 15 * time.Second
	containerdBinary          = "docker-containerd"
	containerdPidFilename     = "docker-containerd.pid"
	containerdSockFilename    = "docker-containerd.sock"
	eventTimestampFilename    = "event.ts"
)

type remote struct {
	sync.RWMutex
	apiClient     containerd.APIClient
	daemonPid     int
	stateDir      string
	rpcAddr       string
	startDaemon   bool
	closeManually bool
	debugLog      bool
	rpcConn       *grpc.ClientConn
	clients       []*client
	eventTsPath   string
	pastEvents    map[string]*containerd.Event
	runtimeArgs   []string
}

// New creates a fresh instance of libcontainerd remote.
func New(stateDir string, options ...RemoteOption) (_ Remote, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Failed to connect to containerd. Please make sure containerd is installed in your PATH or you have specificed the correct address. Got error: %v", err)
		}
	}()
	r := &remote{
		stateDir:    stateDir,
		daemonPid:   -1,
		eventTsPath: filepath.Join(stateDir, eventTimestampFilename),
		pastEvents:  make(map[string]*containerd.Event),
	}
	for _, option := range options {
		if err := option.Apply(r); err != nil {
			return nil, err
		}
	}

	if err := sysinfo.MkdirAll(stateDir, 0700); err != nil {
		return nil, err
	}

	if r.rpcAddr == "" {
		r.rpcAddr = filepath.Join(stateDir, containerdSockFilename)
	}

	if r.startDaemon {
		if err := r.runContainerdDaemon(); err != nil {
			return nil, err
		}
	}

	if err := r.startEventsMonitor(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *remote) Cleanup() {
}

func (r *remote) Client(b Backend) (Client, error) {
	c := &client{
		clientCommon: clientCommon{
			backend:    b,
			containers: make(map[string]*container),
			locker:     locker.New(),
		},
		remote:        r,
		exitNotifiers: make(map[string]*exitNotifier),
	}

	r.Lock()
	r.clients = append(r.clients, c)
	r.Unlock()
	return c, nil
}

func (r *remote) updateEventTimestamp(t time.Time) {
	f, err := os.OpenFile(r.eventTsPath, syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		logrus.Warnf("libcontainerd: failed to open event timestamp file: %v", err)
		return
	}

	b, err := t.MarshalText()
	if err != nil {
		logrus.Warnf("libcontainerd: failed to encode timestamp: %v", err)
		return
	}

	n, err := f.Write(b)
	if err != nil || n != len(b) {
		logrus.Warnf("libcontainerd: failed to update event timestamp file: %v", err)
		f.Truncate(0)
		return
	}

}

func (r *remote) getLastEventTimestamp() int64 {
	t := time.Now()

	fi, err := os.Stat(r.eventTsPath)
	if os.IsNotExist(err) || fi.Size() == 0 {
		return t.Unix()
	}

	f, err := os.Open(r.eventTsPath)
	defer f.Close()
	if err != nil {
		logrus.Warn("libcontainerd: Unable to access last event ts: %v", err)
		return t.Unix()
	}

	b := make([]byte, fi.Size())
	n, err := f.Read(b)
	if err != nil || n != len(b) {
		logrus.Warn("libcontainerd: Unable to read last event ts: %v", err)
		return t.Unix()
	}

	t.UnmarshalText(b)

	return t.Unix()
}

func (r *remote) startEventsMonitor() error {
	// First, get past events
	er := &containerd.EventsRequest{
		Timestamp: uint64(r.getLastEventTimestamp()),
	}
	events, err := r.apiClient.Events(context.Background(), er)
	if err != nil {
		return err
	}
	go r.handleEventStream(events)
	return nil
}

func (r *remote) handleEventStream(events containerd.API_EventsClient) {
	live := false
	for {
		e, err := events.Recv()
		if err != nil {
			if grpc.ErrorDesc(err) == transport.ErrConnClosing.Desc &&
				r.closeManually {
				// ignore error if grpc remote connection is closed manually
				return
			}
			logrus.Errorf("failed to receive event from containerd: %v", err)
			go r.startEventsMonitor()
			return
		}

		if live == false {
			logrus.Debugf("received past containerd event: %#v", e)

			// Pause/Resume events should never happens after exit one
			switch e.Type {
			case StateExit:
				r.pastEvents[e.Id] = e
			case StatePause:
				r.pastEvents[e.Id] = e
			case StateResume:
				r.pastEvents[e.Id] = e
			case stateLive:
				live = true
				r.updateEventTimestamp(time.Unix(int64(e.Timestamp), 0))
			}
		} else {
			logrus.Debugf("received containerd event: %#v", e)

			var container *container
			var c *client
			r.RLock()
			for _, c = range r.clients {
				container, err = c.getContainer(e.Id)
				if err == nil {
					break
				}
			}
			r.RUnlock()
			if container == nil {
				logrus.Errorf("no state for container: %q", err)
				continue
			}

			if err := container.handleEvent(e); err != nil {
				logrus.Errorf("error processing state change for %s: %v", e.Id, err)
			}

			r.updateEventTimestamp(time.Unix(int64(e.Timestamp), 0))
		}
	}
}

func (r *remote) runContainerdDaemon() error {
	var err error
	r.apiClient, err = newBridge(r.stateDir, 10, "docker-runc", r.runtimeArgs)
	return err
}

// WithRemoteAddr sets the external containerd socket to connect to.
func WithRemoteAddr(addr string) RemoteOption {
	return rpcAddr(addr)
}

type rpcAddr string

func (a rpcAddr) Apply(r Remote) error {
	if remote, ok := r.(*remote); ok {
		remote.rpcAddr = string(a)
		return nil
	}
	return fmt.Errorf("WithRemoteAddr option not supported for this remote")
}

// WithRuntimeArgs sets the list of runtime args passed to containerd
func WithRuntimeArgs(args []string) RemoteOption {
	return runtimeArgs(args)
}

type runtimeArgs []string

func (rt runtimeArgs) Apply(r Remote) error {
	if remote, ok := r.(*remote); ok {
		remote.runtimeArgs = rt
		return nil
	}
	return fmt.Errorf("WithRuntimeArgs option not supported for this remote")
}

// WithStartDaemon defines if libcontainerd should also run containerd daemon.
func WithStartDaemon(start bool) RemoteOption {
	return startDaemon(start)
}

type startDaemon bool

func (s startDaemon) Apply(r Remote) error {
	if remote, ok := r.(*remote); ok {
		remote.startDaemon = bool(s)
		return nil
	}
	return fmt.Errorf("WithStartDaemon option not supported for this remote")
}

// WithDebugLog defines if containerd debug logs will be enabled for daemon.
func WithDebugLog(debug bool) RemoteOption {
	return debugLog(debug)
}

type debugLog bool

func (d debugLog) Apply(r Remote) error {
	if remote, ok := r.(*remote); ok {
		remote.debugLog = bool(d)
		return nil
	}
	return fmt.Errorf("WithDebugLog option not supported for this remote")
}
