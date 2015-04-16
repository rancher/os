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
		BootstrapDocker: DockerConfig{
			Args: []string{
				"docker",
				"-d",
				"-s", "overlay",
				"-b", "none",
				"--restart=false",
				"-g", "/var/lib/system-docker",
				"-G", "root",
				"-H", DOCKER_SYSTEM_HOST,
			},
		},
		SystemDocker: DockerConfig{
			Args: []string{
				"docker",
				"-d",
				"--log-driver", "syslog",
				"-s", "overlay",
				"-b", "docker-sys",
				"--fixed-cidr", "172.18.42.1/16",
				"--restart=false",
				"-g", "/var/lib/system-docker",
				"-G", "root",
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
					"/lib/modules:/lib/modules",
					"/lib/firmware:/lib/firmware",
				},
				Image:     "udev",
				LogDriver: "json-file",
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
				VolumesFrom: []string{
					"system-volumes",
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
					"/dev:/host/dev",
					"/var/lib/rancher/conf:/var/lib/rancher/conf",
					"/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt.rancher",
					"/lib/modules:/lib/modules",
					"/lib/firmware:/lib/firmware",
					"/var/run:/var/run",
					"/var/log:/var/log",
				},
				LogDriver: "json-file",
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
					"/init:/usr/sbin/wait-for-docker:ro",
					"/lib/modules:/lib/modules",
					"/usr/bin/docker:/usr/bin/docker:ro",
				},
				LogDriver: "json-file",
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
				LogDriver: "json-file",
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
				LogDriver: "json-file",
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
				LogDriver: "json-file",
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
				Image:      "network",
				Privileged: true,
				Net:        "host",
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
				VolumesFrom: []string{
					"system-volumes",
				},
				LogDriver: "json-file",
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
			"userdockerwait": {
				Image: "userdockerwait",
				Net:   "host",
				Labels: []string{
					"io.rancher.os.detach=false",
				},
				Links: []string{
					"userdocker",
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
		Services: map[string]bool{
			"ubuntu-console": false,
		},
		BundledServices: map[string]Config{
			"ubuntu-console": {
				SystemContainers: map[string]*project.ServiceConfig{
					"console": {
						Image:      "rancher/ubuntuconsole:" + IMAGE_VERSION,
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
