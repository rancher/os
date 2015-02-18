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
		SystemDockerArgs: []string{"docker", "-d", "-s", "overlay", "-b", "none"},
		Modules:          []string{},
		SystemContainers: []ContainerConfig{
			{
				Cmd: "--name=system-state " +
					"--net=none " +
					"--read-only " +
					"-v=/var/lib/rancher/etc:/var/lib/rancher/etc " +
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
				Cmd: "--name=network " +
					"--cap-add=NET_ADMIN " +
					"--net=host " +
					"--rm " +
					"network",
			},
			{
				Cmd: "--name=userdocker " +
					"-d " +
					"--restart=always " +
					"--pid=host " +
					"--net=host " +
					"--privileged " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"--volumes-from=system-state " +
					"userdocker",
			},
			{
				Cmd: "--name=console " +
					"-d " +
					"--rm " +
					"--privileged " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"-v=/init:/usr/bin/system-docker:ro " +
					"-v=/init:/usr/bin/respawn:ro " +
					"-v=/var/run/docker.sock:/var/run/system-docker.sock:ro " +
					"-v=/init:/sbin/poweroff:ro " +
					"-v=/init:/sbin/reboot:ro " +
					"-v=/init:/sbin/halt:ro " +
					"-v=/init:/sbin/tlsconf:ro " +
					"-v=/init:/usr/bin/rancherctl:ro " +
					"--volumes-from=system-state " +
					"--net=host " +
					"--pid=host " +
					"console",
			},
			{
				Cmd: "--name=ntp " +
					"-d " +
					"--privileged " +
					"--net=host " +
					"ntp",
			},
		},
		RescueContainer: &ContainerConfig{
			Cmd: "--name=rescue " +
				"-d " +
				"--rm " +
				"--privileged " +
				"-v=/lib/modules:/lib/modules:ro " +
				"-v=/usr/bin/docker:/usr/bin/docker:ro " +
				"-v=/init:/usr/bin/system-docker:ro " +
				"-v=/init:/usr/bin/respawn:ro " +
				"-v=/var/run/docker.sock:/var/run/system-docker.sock:ro " +
				"--net=host " +
				"--pid=host " +
				"rescue",
		},
	}
}
