---
title: Configuring Docker in RancherOS
layout: os-default

---

## Configuring Docker or System Docker

In RancherOS, you can configure System Docker and Docker daemons by using [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config). 

### Configuring Docker

In your cloud-config, Docker configuration is located under the `rancher.docker` key. 

```yaml
#cloud-config
rancher:
  docker:
    tls: true
    tls_args: [--tlsverify, --tlscacert=ca.pem, --tlscert=server-cert.pem, --tlskey=server-key.pem, '-H=0.0.0.0:2376']
    storage_driver: overlay
```    

You can also customize Docker after it's been started using `ros config`.

```
$ sudo ros config set rancher.docker.storage_driver overlay
```

#### Valid Keys for Docker

Many of the standard Docker daemon arguments can be placed under the `rancher.docker` key. The command needed to start the Docker daemon will be generated based on these arguments. The following arguments are currently supported.

Key | Value
---|---
`bridge` | String
`config_file` | String
`containerd` | String
`debug` | Boolean
`exec_root` | String
`group` | String
`graph` | String
`host` | List
`insecure_registry` | List
`live_restore` | Boolean
`log_driver` | String
`log_opts` | Map where keys and values are strings
`pid_file` | String
`registry_mirror` | String
`restart` | Boolean
`selinux_enabled` | Boolean
`storage_driver` | String
`userland_proxy` | Boolean

In addition to the standard daemon arguments, there are a few fields specific to RancherOS.

Key | Value | Default | Description
---|---|---| ---
`extra_args` | List of Strings | `[]` | Arbitrary daemon arguments, appended to the generated command
`environment` | List of Strings | `[]` | 
`tls` | Boolean | `false` | When [setting up TLS]({{site.baseurl}}/os/configuration/setting-up-docker-tls/), this key needs to be set to true.
`tls_args` | List of Strings (used only if `tls: true`) | `[]` | 
`server_key` | String (used only if `tls: true`)| `""` | PEM encoded server TLS key. 
`server_cert` | String (used only if `tls: true`) | `""` | PEM encoded server TLS certificate.
`ca_key` | String (used only if `tls: true`) | `""` | PEM encoded CA TLS key. 
`storage_context` | String | `console` | Specifies the name of the system container in whose context to run the Docker daemon process.

### Configuring System Docker

In your cloud-config, System Docker configuration is located under the `rancher.system_docker` key. 

```yaml
#cloud-config
rancher:
  system_docker:
    storage_driver: overlay
```

#### Valid Keys for System Docker

All daemon arguments shown in the first table are also available to System Docker. The following are also supported.

Key | Value | Default | Description
---|---|---| ---
`extra_args` | List of Strings | `[]` | Arbitrary daemon arguments, appended to the generated command
`environment` | List of Strings (optional) | `[]` | 
