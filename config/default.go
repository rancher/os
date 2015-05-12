package config

import (
	"github.com/rancherio/rancher-compose/librcompose/project"
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
		BootstrapContainers: map[string]*project.ServiceConfig{
			"udev": {
				Net:        "host",
				Privileged: true,
				Labels: project.NewSliceorMap(map[string]string{
					DETACH: "false",
					SCOPE:  SYSTEM,
				}),
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
				Restart:    "always",
				Net:        "host",
				Privileged: true,
				Labels: project.NewSliceorMap(map[string]string{
					DETACH: "true",
					SCOPE:  SYSTEM,
				}),
				Environment: project.NewMaporslice([]string{
					"DAEMON=true",
				}),
				VolumesFrom: []string{
					"system-volumes",
				},
			},
			"system-volumes": {
				Image:      "state",
				Net:        "none",
				ReadOnly:   true,
				Privileged: true,
				Labels: project.NewSliceorMap(map[string]string{
					CREATE_ONLY: "true",
					SCOPE:       SYSTEM,
				}),
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
				Labels: project.NewSliceorMap(map[string]string{
					CREATE_ONLY: "true",
					SCOPE:       SYSTEM,
				}),
				Volumes: []string{
					"/init:/sbin/halt:ro",
					"/init:/sbin/poweroff:ro",
					"/init:/sbin/reboot:ro",
					"/init:/sbin/shutdown:ro",
					"/init:/sbin/netconf:ro",
					"/init:/usr/bin/cloud-init:ro",
					"/init:/usr/bin/rancherctl:ro", // deprecated, use `ros` instead
					"/init:/usr/bin/ros:ro",
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
				Labels: project.NewSliceorMap(map[string]string{
					CREATE_ONLY: "true",
					SCOPE:       SYSTEM,
				}),
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
				Labels: project.NewSliceorMap(map[string]string{
					CREATE_ONLY: "true",
					SCOPE:       SYSTEM,
				}),
				Volumes: []string{
					"/var/lib/rancher/conf:/var/lib/rancher/conf",
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
				Labels: project.NewSliceorMap(map[string]string{
					CREATE_ONLY: "true",
					SCOPE:       SYSTEM,
				}),
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
				Labels: project.NewSliceorMap(map[string]string{
					RELOAD_CONFIG: "true",
					DETACH:        "false",
					SCOPE:         SYSTEM,
				}),
				Environment: project.NewMaporslice([]string{
					"CLOUD_INIT_NETWORK=false",
				}),
				VolumesFrom: []string{
					"command-volumes",
					"system-volumes",
				},
			},
			"network": {
				Image:      "network",
				Privileged: true,
				Net:        "host",
				Labels: project.NewSliceorMap(map[string]string{
					DETACH: "false",
					SCOPE:  SYSTEM,
				}),
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
				Labels: project.NewSliceorMap(map[string]string{
					RELOAD_CONFIG: "true",
					DETACH:        "false",
					SCOPE:         SYSTEM,
				}),
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
				Restart:    "always",
				Privileged: true,
				Net:        "host",
				Labels: project.NewSliceorMap(map[string]string{
					SCOPE: SYSTEM,
				}),
				Links: []string{
					"cloud-init",
					"network",
				},
			},
			"syslog": {
				Image:      "syslog",
				Restart:    "always",
				Privileged: true,
				Net:        "host",
				Labels: project.NewSliceorMap(map[string]string{
					SCOPE: SYSTEM,
				}),
				VolumesFrom: []string{
					"system-volumes",
				},
				LogDriver: "json-file",
			},
			"docker": {
				Image:      "docker",
				Restart:    "always",
				Privileged: true,
				Pid:        "host",
				Ipc:        "host",
				Net:        "host",
				Labels: project.NewSliceorMap(map[string]string{
					SCOPE: SYSTEM,
				}),
				Links: []string{
					"network",
				},
				VolumesFrom: []string{
					"all-volumes",
				},
			},
			"dockerwait": {
				Image: "dockerwait",
				Net:   "host",
				Labels: project.NewSliceorMap(map[string]string{
					DETACH: "false",
					SCOPE:  SYSTEM,
				}),
				Links: []string{
					"docker",
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
				Labels: project.NewSliceorMap(map[string]string{
					SCOPE: SYSTEM,
				}),
				VolumesFrom: []string{
					"all-volumes",
				},
				Restart: "always",
				Pid:     "host",
				Ipc:     "host",
				Net:     "host",
			},
			"acpid": {
				Image:      "acpid",
				Privileged: true,
				Labels: project.NewSliceorMap(map[string]string{
					SCOPE: SYSTEM,
				}),
				VolumesFrom: []string{
					"command-volumes",
					"system-volumes",
				},
				Net: "host",
			},
		},
		ServicesInclude: map[string]bool{
			"ubuntu-console": false,
		},
		Repositories: map[string]Repository{
			"core": Repository{
				Url: "https://raw.githubusercontent.com/rancherio/os-services/master",
			},
		},
		Services: map[string]*project.ServiceConfig{},
	}
}
