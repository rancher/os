package control

import (
	"fmt"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/rancher/os/config"
)

func selinuxCommand() cli.Command {
	app := cli.Command{}
	app.Name = "selinux"
	app.Usage = "Launch SELinux tools container."
	app.Action = func(c *cli.Context) error {
		argv := []string{"system-docker", "run", "-it", "--privileged", "--rm",
			"--net", "host", "--pid", "host", "--ipc", "host",
			"-v", "/usr/bin/docker:/usr/bin/docker.dist:ro",
			"-v", "/usr/bin/ros:/usr/bin/dockerlaunch:ro",
			"-v", "/usr/bin/ros:/usr/bin/user-docker:ro",
			"-v", "/usr/bin/ros:/usr/bin/system-docker:ro",
			"-v", "/usr/bin/ros:/sbin/poweroff:ro",
			"-v", "/usr/bin/ros:/sbin/reboot:ro",
			"-v", "/usr/bin/ros:/sbin/halt:ro",
			"-v", "/usr/bin/ros:/sbin/shutdown:ro",
			"-v", "/usr/bin/ros:/usr/bin/respawn:ro",
			"-v", "/usr/bin/ros:/usr/bin/ros:ro",
			"-v", "/usr/bin/ros:/usr/bin/cloud-init:ro",
			"-v", "/usr/bin/ros:/usr/sbin/netconf:ro",
			"-v", "/usr/bin/ros:/usr/sbin/wait-for-network:ro",
			"-v", "/usr/bin/ros:/usr/sbin/wait-for-docker:ro",
			"-v", "/var/lib/docker:/var/lib/docker",
			"-v", "/var/lib/rkt:/var/lib/rkt",
			"-v", "/dev:/host/dev",
			"-v", "/etc/docker:/etc/docker",
			"-v", "/etc/hosts:/etc/hosts",
			"-v", "/etc/resolv.conf:/etc/resolv.conf",
			"-v", "/etc/rkt:/etc/rkt",
			"-v", "/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt.rancher",
			"-v", "/lib/firmware:/lib/firmware",
			"-v", "/lib/modules:/lib/modules",
			"-v", "/run:/run",
			"-v", "/usr/share/ros:/usr/share/ros",
			"-v", "/var/lib/rancher/conf:/var/lib/rancher/conf",
			"-v", "/var/lib/rancher:/var/lib/rancher",
			"-v", "/var/log:/var/log",
			"-v", "/var/run:/var/run",
			"-v", "/home:/home",
			"-v", "/opt:/opt",
			"-v", "/etc/selinux:/etc/selinux",
			"-v", "/var/lib/selinux:/var/lib/selinux",
			"-v", "/usr/share/selinux:/usr/share/selinux",
			fmt.Sprintf("%s/os-selinuxtools:%s%s", config.OS_REPO, config.VERSION, config.SUFFIX), "bash"}
		syscall.Exec("/bin/system-docker", argv, []string{})
		return nil
	}

	return app
}
