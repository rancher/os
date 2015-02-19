package config

func NewConfig() *Config {
	return &Config{
		Debug: DEBUG,
		Dns: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
		State: ConfigState{
			Required: false,
			Dev:      "LABEL=RANCHER_STATE",
			FsType:   "auto",
		},
		SystemDockerArgs: []string{"docker", "-d", "-s", "overlay", "-b", "none", "--restart=false", "-H", DOCKER_SYSTEM_HOST},
		Modules:          []string{},
		Userdocker: UserDockerInfo{
			UseTLS: true,
		},
		CloudConfig: []string{},
		SystemContainers: []ContainerConfig{
			{
				Cmd: "--name=system-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/var/lib/rancher/conf:/var/lib/rancher/conf " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/var/run:/var/run " +
					"state",
			},
			{
				Cmd: "--name=console-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/init:/sbin/halt:ro " +
					"-v=/init:/sbin/poweroff:ro " +
					"-v=/init:/sbin/reboot:ro " +
					"-v=/init:/sbin/tlsconf:ro " +
					"-v=/init:/usr/bin/tlsconf:ro " +
					"-v=/init:/usr/bin/rancherctl:ro " +
					"-v=/init:/usr/bin/respawn:ro " +
					"-v=/init:/usr/bin/system-docker:ro " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"-v=/var/lib/rancher/state/home:/home " +
					"-v=/var/lib/rancher/state/opt:/opt " +
					"state",
			},
			{
				Cmd: "--name=udev " +
					"--net=none " +
					"--privileged " +
					"--rm " +
					"-v=/dev:/host/dev " +
					"-v=/lib/modules:/lib/modules:ro " +
					"udev",
			},
			{
				Cmd: "--name=cloudconfig " +
					"--net=host " +
					"-v=/init:/usr/bin/rancherctl:ro " +
					"-v=/init:/usr/bin/cloudinit:ro " +
					"cloudconfig",
			},
			{
				Cmd: "--name=network " +
					"--cap-add=NET_ADMIN " +
					"--net=host " +
					"--rm " +
					"network",
			},
			{
				Cmd: "--name=ntp " +
					"--rm " +
					"-d " +
					"--privileged " +
					"--net=host " +
					"ntp",
			},
			{
				Cmd: "--name=userdocker " +
					"-d " +
					"--rm " +
					"--restart=always " +
					"--ipc=host " +
					"--pid=host " +
					"--net=host " +
					"--privileged " +
					"--volumes-from=console-volumes " +
					"--volumes-from=system-volumes " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"-v=/init:/usr/bin/tlsconf:ro " +
					"-v=/init:/usr/sbin/rancherctl:ro " +
					"-v=/var/lib/rancher/state/docker:/var/lib/docker " +
					"userdocker",
			},
			{
				Cmd: "--name=console " +
					"-d " +
					"--rm " +
					"--privileged " +
					"--volumes-from=console-volumes " +
					"--volumes-from=system-volumes " +
					"--ipc=host " +
					"--net=host " +
					"--pid=host " +
					"console",
			},
			{
				Cmd: "--name=syslog " +
				     	"-d " +
				     	"--privileged " +
				    	"--net=host " +
					"syslog",
			},
		},
		RescueContainer: &ContainerConfig{
			Cmd: "--name=rescue " +
				"-d " +
				"--rm " +
				"--privileged " +
				"--volumes-from=console-volumes " +
				"--volumes-from=system-volumes " +
				"--ipc=host " +
				"--net=host " +
				"--pid=host " +
				"rescue",
		},
	}
}
