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
		SystemContainers: []ContainerConfig{
			{
				Id: "system-volumes",
				Cmd: "--name=system-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/var/lib/rancher/conf:/var/lib/rancher/conf " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/var/run:/var/run " +
					"-v=/var/log:/var/log " +
					"state",
			},
			{
				Id: "command-volumes",
				Cmd: "--name=command-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/init:/sbin/halt:ro " +
					"-v=/init:/sbin/poweroff:ro " +
					"-v=/init:/sbin/reboot:ro " +
					"-v=/init:/usr/bin/cloud-init:ro " +
					"-v=/init:/usr/bin/tlsconf:ro " +
					"-v=/init:/usr/bin/rancherctl:ro " +
					"-v=/init:/usr/bin/respawn:ro " +
					"-v=/init:/usr/bin/system-docker:ro " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"state",
			},
			{
				Id: "user-volumes",
				Cmd: "--name=user-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/var/lib/rancher/state/home:/home " +
					"-v=/var/lib/rancher/state/opt:/opt " +
					"state",
			},
			{
				Id: "udev",
				Cmd: "--name=udev " +
					"--net=none " +
					"--privileged " +
					"--rm " +
					"-v=/dev:/host/dev " +
					"-v=/lib/modules:/lib/modules:ro " +
					"udev",
			},
			{
				Id: "cloud-init",
				Cmd: "--name=cloud-init " +
					"--rm " +
					"--net=host " +
					"--volumes-from=command-volumes " +
					"cloudinit",
				ReloadConfig: true,
			},
			{
				Id: "network",
				Cmd: "--name=network " +
					"--cap-add=NET_ADMIN " +
					"--net=host " +
					"--rm " +
					"network",
			},
			{
				Id: "ntp",
				Cmd: "--name=ntp " +
					"--rm " +
					"-d " +
					"--privileged " +
					"--net=host " +
					"ntp",
			},
			{
				Id: "syslog",
				Cmd: "--name=syslog " +
					"-d " +
					"--rm " +
					"--privileged " +
					"--net=host " +
					"--ipc=host " +
					"--pid=host " +
					"--volumes-from=system-volumes " +
					"syslog",
			},
			{
				Id: "userdocker",
				Cmd: "--name=userdocker " +
					"-d " +
					"--rm " +
					"--restart=always " +
					"--ipc=host " +
					"--pid=host " +
					"--net=host " +
					"--privileged " +
					"--volumes-from=command-volumes " +
					"--volumes-from=user-volumes " +
					"--volumes-from=system-volumes " +
					"-v=/var/lib/rancher/state/docker:/var/lib/docker " +
					"userdocker",
			},
			{
				Id: "console",
				Cmd: "--name=console " +
					"-d " +
					"--rm " +
					"--privileged " +
					"--volumes-from=command-volumes " +
					"--volumes-from=user-volumes " +
					"--volumes-from=system-volumes " +
					"--restart=always " +
					"--ipc=host " +
					"--net=host " +
					"--pid=host " +
					"console",
			},
		},
		EnabledAddons: []string{},
		Addons: map[string]Config{
			"ubuntu-console": {
				SystemContainers: []ContainerConfig{
					{
						Id: "console",
						Cmd: "--name=ubuntu-console " +
							"-d " +
							"--rm " +
							"--privileged " +
							"--volumes-from=command-volumes " +
							"--volumes-from=user-volumes " +
							"--volumes-from=system-volumes " +
							"--restart=always " +
							"--ipc=host " +
							"--net=host " +
							"--pid=host " +
							"rancher/ubuntuconsole",
					},
				},
			},
		},
		RescueContainer: &ContainerConfig{
			Id: "console",
			Cmd: "--name=rescue " +
				"-d " +
				"--rm " +
				"--privileged " +
				"--volumes-from=console-volumes " +
				"--volumes-from=user-volumes " +
				"--volumes-from=system-volumes " +
				"--restart=always " +
				"--ipc=host " +
				"--net=host " +
				"--pid=host " +
				"rescue",
		},
	}
}
