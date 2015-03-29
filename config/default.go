package config

import (
	"github.com/rancherio/rancher-compose/project"
)

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
				"eth0": {
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
		BootstrapContainers: map[string]*project.ServiceConfig{
			"udev": {
				Net:        "host",
				Privileged: true,
				Labels: []string{
					DETACH + "=false",
				},
				Volumes: []string{
					"/dev:/host/dev",
					"/lib/modules:/lib/modules:ro",
					"/lib/firmware:/lib/firmware:ro",
				},
				Image: "udev",
			},
		},
		SystemContainers: map[string]*project.ServiceConfig{
			"udev": {
				Image:      "udev",
				Net:        "host",
				Privileged: true,
				Labels: []string{
					DETACH + "=true",
				},
				Environment: []string{
					"DAEMON=true",
				},
				Volumes: []string{
					"/dev:/host/dev",
					"/lib/modules:/lib/modules:ro",
					"/lib/firmware:/lib/firmware:ro",
				},
			},
			"system-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: []string{
					CREATE_ONLY + "=true",
				},
				Volumes: []string{
					"/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt",
					"/var/lib/rancher/conf:/var/lib/rancher/conf",
					"/lib/modules:/lib/modules:ro",
					"/lib/firmware:/lib/firmware:ro",
					"/var/run:/var/run",
					"/var/log:/var/log",
				},
			},
			"command-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: []string{
					CREATE_ONLY + "=true",
				},
				Volumes: []string{
					"/init:/sbin/halt:ro",
					"/init:/sbin/poweroff:ro",
					"/init:/sbin/reboot:ro",
					"/init:/sbin/shutdown:ro",
					"/init:/sbin/netconf:ro",
					"/init:/usr/bin/cloud-init:ro",
					"/init:/usr/bin/rancherctl:ro",
					"/init:/usr/bin/respawn:ro",
					"/init:/usr/bin/system-docker:ro",
					"/lib/modules:/lib/modules:ro",
					"/usr/bin/docker:/usr/bin/docker:ro",
				},
			},
			"user-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: []string{
					CREATE_ONLY + "=true",
				},
				Volumes: []string{
					"/home:/home",
					"/opt:/opt",
				},
			},
			"docker-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: []string{
					CREATE_ONLY + "=true",
				},
				Volumes: []string{
					"/var/lib/rancher:/var/lib/rancher",
					"/var/lib/docker:/var/lib/docker",
					"/var/lib/system-docker:/var/lib/system-docker",
				},
			},
			"all-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: []string{
					CREATE_ONLY + "=true",
				},
				VolumesFrom: []string{
					"docker-volumes",
					"command-volumes",
					"user-volumes",
					"system-volumes",
				},
			},
			"cloud-init-pre": {
				Image:      "cloudinit",
				Privileged: true,
				Net:        "host",
				Labels: []string{
					RELOAD_CONFIG + "=true",
					DETACH + "=false",
				},
				Environment: []string{
					"CLOUD_INIT_NETWORK=false",
				},
				VolumesFrom: []string{
					"command-volumes",
					"system-volumes",
				},
			},
			"network": {
				Image: "network",
				CapAdd: []string{
					"NET_ADMIN",
				},
				Net: "host",
				Labels: []string{
					DETACH + "=false",
				},
				Links: []string{
					"cloud-init-pre",
				},
				VolumesFrom: []string{
					"command-volumes",
					"system-volumes",
				},
			},
			"cloud-init": {
				Image:      "cloudinit",
				Privileged: true,
				Labels: []string{
					RELOAD_CONFIG + "=true",
					DETACH + "=false",
				},
				Net: "host",
				Links: []string{
					"cloud-init-pre",
					"network",
				},
				VolumesFrom: []string{
					"command-volumes",
					"system-volumes",
				},
			},
			"ntp": {
				Image:      "ntp",
				Privileged: true,
				Net:        "host",
				Links: []string{
					"cloud-init",
					"network",
				},
			},
			"syslog": {
				Image:      "syslog",
				Privileged: true,
				Net:        "host",
				Links: []string{
					"cloud-init",
					"network",
				},
				VolumesFrom: []string{
					"system-volumes",
				},
			},
			"userdocker": {
				Image:      "userdocker",
				Privileged: true,
				Pid:        "host",
				Ipc:        "host",
				Net:        "host",
				Links: []string{
					"network",
				},
				VolumesFrom: []string{
					"all-volumes",
				},
			},
			"console": {
				Image:      "console",
				Privileged: true,
				Links: []string{
					"cloud-init",
				},
				VolumesFrom: []string{
					"all-volumes",
				},
				Restart: "always",
				Pid:     "host",
				Ipc:     "host",
				Net:     "host",
			},
		},
		EnabledAddons: []string{},
		Addons: map[string]Config{
			"ubuntu-console": {
				SystemContainers: map[string]*project.ServiceConfig{
					"console": {
						Image:      "rancher/ubuntuconsole:" + VERSION,
						Privileged: true,
						Labels: []string{
							DETACH + "=true",
						},
						Links: []string{
							"cloud-init",
						},
						VolumesFrom: []string{
							"all-volumes",
						},
						Restart: "always",
						Pid:     "host",
						Ipc:     "host",
						Net:     "host",
					},
				},
			},
		},
	}
}
