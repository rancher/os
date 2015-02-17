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
				Cmd: []string{
					"--name", "system-state",
					"--net", "none",
					"--read-only",
					"state",
				},
			},
			{
				Cmd: []string{
					"--name", "udev",
					"--net", "none",
					"--privileged",
					"--rm",
					"--volume", "/dev:/host/dev",
					"--volume", "/lib/modules:/lib/modules:ro",
					"udev",
				},
			},
			{
				Cmd: []string{
					"--name", "network",
					"--cap-add", "NET_ADMIN",
					"--net", "host",
					"--rm",
					"network",
				},
			},
			{
				Cmd: []string{
					"--name", "userdocker",
					"-d",
					"--restart", "always",
					"--pid", "host",
					"--net", "host",
					"--privileged",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volumes-from", "system-state",
					"userdocker",
				},
			},
			{
				Cmd: []string{
					"--name", "console",
					"-d",
					"--rm",
					"--privileged",
					"--volume", "/lib/modules:/lib/modules:ro",
					"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
					"--volume", "/init:/usr/bin/system-docker:ro",
					"--volume", "/init:/usr/bin/respawn:ro",
					"--volume", "/var/run/docker.sock:/var/run/system-docker.sock:ro",
					"--volume", "/init:/sbin/poweroff:ro",
					"--volume", "/init:/sbin/reboot:ro",
					"--volume", "/init:/sbin/halt:ro",
					"--volumes-from", "system-state",
					"--net", "host",
					"--pid", "host",
					"console",
				},
			},
			{
				Cmd: []string{
					"--name", "ntp",
					"-d",
					"--privileged",
					"--net", "host",
					"ntp",
				},
			},
		},
		RescueContainer: ContainerConfig{
			Cmd: []string{
				"--name", "rescue",
				"-d",
				"--rm",
				"--privileged",
				"--volume", "/lib/modules:/lib/modules:ro",
				"--volume", "/usr/bin/docker:/usr/bin/docker:ro",
				"--volume", "/init:/usr/bin/system-docker:ro",
				"--volume", "/init:/usr/bin/respawn:ro",
				"--volume", "/var/run/docker.sock:/var/run/system-docker.sock:ro",
				"--net", "host",
				"--pid", "host",
				"rescue",
			},
		},
	}
}
