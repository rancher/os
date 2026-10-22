package control

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/burmilla/os/cmd/respawn"
	"github.com/burmilla/os/config"

	"github.com/codegangsta/cli"
)

const sshdPidFile = "/run/sshd.pid"

func sshdSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "status",
			Usage:  "print whether sshd of the console is enabled or disabled",
			Action: sshdStatus,
		},
		{
			Name:  "disable",
			Usage: "stop sshd of the console until it is enabled again or the system reboots",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "if-idle",
					Usage: "only disable when no SSH connection is established",
				},
			},
			Action: sshdDisable,
		},
		{
			Name:   "enable",
			Usage:  "start sshd of the console again",
			Action: sshdEnable,
		},
	}
}

func sshdMarker() string {
	return filepath.Join(respawn.DisabledDir, "sshd")
}

func sshdStatus(c *cli.Context) error {
	if respawn.Disabled("sshd") {
		fmt.Println("disabled")
	} else {
		fmt.Println("enabled")
	}
	return nil
}

func sshdDisable(c *cli.Context) error {
	if c.Bool("if-idle") {
		port := config.LoadConfig().Rancher.SSH.Port
		if port <= 0 {
			port = 22
		}
		n, err := establishedConnections(port)
		if err != nil {
			return err
		}
		if n > 0 {
			fmt.Printf("not disabled: %d SSH connection(s) established\n", n)
			return nil
		}
	}

	if err := os.MkdirAll(respawn.DisabledDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(sshdMarker(), nil, 0644); err != nil {
		return err
	}
	// Established sessions run in their own processes and are kept, only
	// the listener is stopped. respawn does not start it again while the
	// marker exists.
	if err := stopSshd(); err != nil {
		return err
	}
	fmt.Println("disabled")
	return nil
}

func sshdEnable(c *cli.Context) error {
	if err := os.Remove(sshdMarker()); err != nil && !os.IsNotExist(err) {
		return err
	}
	fmt.Println("enabled")
	return nil
}

func stopSshd() error {
	data, err := os.ReadFile(sshdPidFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 1 {
		return fmt.Errorf("invalid pid in %s", sshdPidFile)
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if err == syscall.ESRCH {
			return nil
		}
		return err
	}
	for i := 0; i < 50; i++ {
		if syscall.Kill(pid, 0) == syscall.ESRCH {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("sshd (pid %d) did not stop", pid)
}

// establishedConnections counts established TCP connections whose local
// port is port.
func establishedConnections(port int) (int, error) {
	n := 0
	for _, file := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		f, err := os.Open(file)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return 0, err
		}
		c, err := countEstablished(bufio.NewScanner(f), port)
		f.Close()
		if err != nil {
			return 0, err
		}
		n += c
	}
	return n, nil
}

func countEstablished(s *bufio.Scanner, port int) (int, error) {
	const established = "01"
	n := 0
	for s.Scan() {
		f := strings.Fields(s.Text())
		// sl local_address rem_address st ...
		if len(f) < 4 || f[3] != established {
			continue
		}
		i := strings.LastIndex(f[1], ":")
		if i < 0 {
			continue
		}
		p, err := strconv.ParseUint(f[1][i+1:], 16, 16)
		if err == nil && int(p) == port {
			n++
		}
	}
	return n, s.Err()
}
