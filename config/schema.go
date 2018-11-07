package config

var schema = `{
	"type": "object",
	"additionalProperties": false,

	"properties": {
		"ssh_authorized_keys": {"$ref": "#/definitions/list_of_strings"},
		"write_files": {
			"type": "array",
			"items": {"$ref": "#/definitions/file_config"}
		},
		"hostname": {"type": "string"},
		"mounts": {"type": "array"},
		"rancher": {"$ref": "#/definitions/rancher_config"},
		"runcmd": {"type": "array"},
		"bootcmd": {"type": "array"}
	},

	"definitions": {
		"rancher_config": {
			"id": "#/definitions/rancher_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"console": {"type": "string"},
				"environment": {"type": "object"},
				"cloud_init_services": {"type": "object"},
				"services": {"type": "object"},
				"bootstrap": {"type": "object"},
				"autoformat": {"type": "object"},
				"bootstrap_docker": {"$ref": "#/definitions/docker_config"},
				"cloud_init": {"$ref": "#/definitions/cloud_init_config"},
				"debug": {"type": "boolean"},
				"rm_usr": {"type": "boolean"},
				"no_sharedroot": {"type": "boolean"},
				"log": {"type": "boolean"},
				"force_console_rebuild": {"type": "boolean"},
				"recovery": {"type": "boolean"},
				"disable": {"$ref": "#/definitions/list_of_strings"},
				"services_include": {"type": "object"},
				"modules": {"$ref": "#/definitions/list_of_strings"},
				"network": {"$ref": "#/definitions/network_config"},
				"repositories": {"type": "object"},
				"ssh": {"$ref": "#/definitions/ssh_config"},
				"state": {"$ref": "#/definitions/state_config"},
				"system_docker": {"$ref": "#/definitions/docker_config"},
				"upgrade": {"$ref": "#/definitions/upgrade_config"},
				"docker": {"$ref": "#/definitions/docker_config"},
				"registry_auths": {"type": "object"},
				"defaults": {"$ref": "#/definitions/defaults_config"},
				"resize_device": {"type": "string"},
				"sysctl": {"type": "object"},
				"restart_services": {"type": "array"},
				"hypervisor_service": {"type": "boolean"},
				"shutdown_timeout": {"type": "integer"},
				"preload_wait": {"type": "boolean"}
			}
		},

		"file_config": {
			"id": "#/definitions/file_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"encoding": {"type": "string"},
				"container": {"type": "string"},
				"content": {"type": "string"},
				"owner": {"type": "string"},
				"path": {"type": "string"},
				"permissions": {"type": "string"}
			}
		},

		"network_config": {
			"id": "#/definitions/network_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"pre_cmds": {"$ref": "#/definitions/list_of_strings"},
				"dhcp_timeout": {"type": "integer"},
				"dns": {"type": "object"},
				"interfaces": {"type": "object"},
				"post_cmds": {"$ref": "#/definitions/list_of_strings"},
				"http_proxy": {"type": "string"},
				"https_proxy": {"type": "string"},
				"no_proxy": {"type": "string"},
				"wifi_networks": {"type": "object"}
			}
		},

		"upgrade_config": {
			"id": "#/definitions/upgrade_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"url": {"type": "string"},
				"image": {"type": "string"},
				"rollback": {"type": "string"}
			}
		},

		"docker_config": {
			"id": "#/definitions/docker_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"engine": {"type": "string"},
				"tls": {"type": "boolean"},
				"tls_args": {"$ref": "#/definitions/list_of_strings"},
				"args": {"$ref": "#/definitions/list_of_strings"},
				"extra_args": {"$ref": "#/definitions/list_of_strings"},
				"server_cert": {"type": "string"},
				"server_key": {"type": "string"},
				"ca_cert": {"type": "string"},
				"ca_key": {"type": "string"},
				"environment": {"$ref": "#/definitions/list_of_strings"},
				"storage_context": {"type": "string"},
				"exec": {"type": ["boolean", "null"]},
				"bridge": {"type": "string"},
				"bip": {"type": "string"},
				"config_file": {"type": "string"},
				"containerd": {"type": "string"},
				"debug": {"type": ["boolean", "null"]},
				"exec_root": {"type": "string"},
				"group": {"type": "string"},
				"graph": {"type": "string"},
				"host": {"type": "array"},
				"live_restore": {"type": ["boolean", "null"]},
				"log_driver": {"type": "string"},
				"log_opts": {"type": "object"},
				"pid_file": {"type": "string"},
				"registry_mirror": {"type": "string"},
				"restart": {"type": ["boolean", "null"]},
				"selinux_enabled": {"type": ["boolean", "null"]},
				"storage_driver": {"type": "string"},
				"userland_proxy": {"type": ["boolean", "null"]},
				"insecure_registry": {"$ref": "#/definitions/list_of_strings"}
			}
		},

		"ssh_config": {
			"id": "#/definitions/ssh_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"keys": {"type": "object"},
				"daemon": {"type": "boolean"},
				"port": {"type": "integer"},
				"listen_address": {"type": "string"}
			}
		},

		"state_config": {
			"id": "#/definitions/state_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"directory": {"type": "string"},
				"fstype": {"type": "string"},
				"dev": {"type": "string"},
				"wait": {"type": "boolean"},
				"required": {"type": "boolean"},
				"autoformat": {"$ref": "#/definitions/list_of_strings"},
				"mdadm_scan": {"type": "boolean"},
				"rngd": {"type": "boolean"},
				"script": {"type": "string"},
				"oem_fstype": {"type": "string"},
				"oem_dev": {"type": "string"},
				"boot_fstype": {"type": "string"},
				"boot_dev": {"type": "string"}
			}
		},

		"cloud_init_config": {
			"id": "#/definitions/cloud_init_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"datasources": {"$ref": "#/definitions/list_of_strings"}
			}
		},

		"defaults_config": {
			"id": "#/definitions/defaults_config",
			"type": "object",
			"additionalProperties": false,

			"properties": {
				"hostname": {"type": "string"},
				"docker": {"type": "object"},
				"network": {"$ref": "#/definitions/network_config"},
				"system_docker_logs": {"type": "string"}
			}
		},

		"list_of_strings": {
			"type": "array",
			"items": {"type": "string"},
			"uniqueItems": true
		}
	}
}

`
