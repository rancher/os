package config

func NewConfig() *Config {
	return &Config{
		Debug: DEBUG,
		State: StateConfig{
			Required: false,
			Dev:      "LABEL=RANCHER_STATE",
			FsType:   "auto",
		},
		SystemDocker: DockerConfig{
			Args: []string{
				"docker",
				"-d",
				"-s",
				"overlay",
				"-b",
				"none",
				"--restart=false",
				"-g", "/var/lib/system-docker",
				"-H", DOCKER_SYSTEM_HOST,
			},
		},
		Modules: []string{},
		UserDocker: DockerConfig{
			TLSArgs: []string{
				"--tlsverify",
				"--tlscacert=ca.pem",
				"--tlscert=server-cert.pem",
				"--tlskey=server-key.pem",
				"-H=0.0.0.0:2376",
			},
			Args: []string{
				"docker",
				"-d",
				"-s", "overlay",
				"-G", "docker",
				"-H", DOCKER_HOST,
			},
		},
		Network: NetworkConfig{
			Dns: DnsConfig{
				Nameservers: []string{"8.8.8.8", "8.8.4.4"},
			},
			Interfaces: map[string]InterfaceConfig{
				"eth*": {
					DHCP: true,
				},
				"lo": {
					Address: "127.0.0.1/8",
				},
			},
		},
		CloudInit: CloudInit{
			Datasources: []string{"configdrive:/media/config-2"},
		},
		Upgrade: UpgradeConfig{
			Url:   "https://releases.rancher.com/os/versions.yml",
			Image: "rancher/os",
		},
		BootstrapContainers: []ContainerConfig{
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
		},
		SystemContainers: []ContainerConfig{
			{
				Id: "udev",
				Cmd: "--name=udev " +
					"--net=none " +
					"--privileged " +
					"--rm " +
					"-v=/dev:/host/dev " +
					"-v=/lib/modules:/lib/modules:ro " +
					"udev",
				CreateOnly: true,
			},
			{
				Id: "system-volumes",
				Cmd: "--name=system-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt " +
					"-v=/var/lib/rancher/conf:/var/lib/rancher/conf " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/var/run:/var/run " +
					"-v=/var/log:/var/log " +
					"state",
				CreateOnly: true,
			},
			{
				Id: "command-volumes",
				Cmd: "--name=command-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/init:/sbin/halt:ro " +
					"-v=/init:/sbin/poweroff:ro " +
					"-v=/init:/sbin/reboot:ro " +
					"-v=/init:/sbin/shutdown:ro " +
					"-v=/init:/sbin/netconf:ro " +
					"-v=/init:/usr/bin/cloud-init:ro " +
					"-v=/init:/usr/bin/rancherctl:ro " +
					"-v=/init:/usr/bin/respawn:ro " +
					"-v=/init:/usr/bin/system-docker:ro " +
					"-v=/lib/modules:/lib/modules:ro " +
					"-v=/usr/bin/docker:/usr/bin/docker:ro " +
					"state",
				CreateOnly: true,
			},
			{
				Id: "user-volumes",
				Cmd: "--name=user-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/home:/home " +
					"-v=/opt:/opt " +
					"state",
				CreateOnly: true,
			},
			{
				Id: "docker-volumes",
				Cmd: "--name=docker-volumes " +
					"--net=none " +
					"--read-only " +
					"-v=/var/lib/rancher:/var/lib/rancher " +
					"-v=/var/lib/docker:/var/lib/docker " +
					"-v=/var/lib/system-docker:/var/lib/system-docker " +
					"state",
				CreateOnly: true,
			},
			{
				Id: "all-volumes",
				Cmd: "--name=all-volumes " +
					"--rm " +
					"--net=none " +
					"--read-only " +
					"--volumes-from=docker-volumes " +
					"--volumes-from=command-volumes " +
					"--volumes-from=user-volumes " +
					"--volumes-from=system-volumes " +
					"state",
				CreateOnly: true,
			},
			{
				Id: "cloud-init-pre",
				Cmd: "--name=cloud-init-pre " +
					"--rm " +
					"--privileged " +
					"--net=host " +
					"-e CLOUD_INIT_NETWORK=false " +
					"--volumes-from=command-volumes " +
					"--volumes-from=system-volumes " +
					"cloudinit",
				ReloadConfig: true,
			},
			{
				Id: "network",
				Cmd: "--name=network " +
					"--rm " +
					"--cap-add=NET_ADMIN " +
					"--net=host " +
					"--volumes-from=command-volumes " +
					"--volumes-from=system-volumes " +
					"network",
			},
			{
				Id: "cloud-init",
				Cmd: "--name=cloud-init " +
					"--rm " +
					"--privileged " +
					"--net=host " +
					"--volumes-from=command-volumes " +
					"--volumes-from=system-volumes " +
					"cloudinit",
				ReloadConfig: true,
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
					"--volumes-from=all-volumes " +
					"userdocker",
			},
			{
				Id: "console",
				Cmd: "--name=console " +
					"-d " +
					"--rm " +
					"--privileged " +
					"--volumes-from=all-volumes " +
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
							"--volumes-from=all-volumes " +
							"--restart=always " +
							"--ipc=host " +
							"--net=host " +
							"--pid=host " +
							"rancher/ubuntuconsole:" + VERSION,
					},
				},
			},
		},
	}
}
