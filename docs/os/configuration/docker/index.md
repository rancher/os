---
title: Configuring Docker in RancherOS


---

## Configuring Docker or System Docker
---

In RancherOS, you can configure System Docker and Docker daemons by using [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config).

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

#### User Docker settings

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
`tls` | Boolean | `false` | When [setting up TLS]({{page.osbaseurl}}/configuration/setting-up-docker-tls/), this key needs to be set to true.
`tls_args` | List of Strings (used only if `tls: true`) | `[]` |
`server_key` | String (used only if `tls: true`)| `""` | PEM encoded server TLS key.
`server_cert` | String (used only if `tls: true`) | `""` | PEM encoded server TLS certificate.
`ca_key` | String (used only if `tls: true`) | `""` | PEM encoded CA TLS key.
`storage_context` | String | `console` | Specifies the name of the system container in whose context to run the Docker daemon process.

#### Example using extra_args for setting MTU

The following example can be used to set MTU on the Docker daemon:

```yaml
#cloud-config
rancher:
  docker:
    extra_args: [--mtu, 1460]
```

### Configuring System Docker

In your cloud-config, System Docker configuration is located under the `rancher.system_docker` key.

```yaml
#cloud-config
rancher:
  system_docker:
    storage_driver: overlay
```

#### System Docker settings

All daemon arguments shown in the first table are also available to System Docker. The following are also supported.

Key | Value | Default | Description
---|---|---| ---
`extra_args` | List of Strings | `[]` | Arbitrary daemon arguments, appended to the generated command
`environment` | List of Strings (optional) | `[]` |

### Using a pull through registry mirror

There are 3 Docker engines that can be configured to use the pull-through Docker Hub registry mirror cache:

```
#cloud-config
rancher:
  bootstrap_docker:
    registry_mirror: "http://10.10.10.23:5555"
  docker:
    registry_mirror: "http://10.10.10.23:5555"
  system_docker:
    registry_mirror: "http://10.10.10.23:5555"
```

`bootstrap_docker` is used to prepare and initial network and pull any cloud-config options that can be used to configure the final network configuration and System-docker - its very unlikely to pull any images.

A successful pull through mirror cache request by System-docker looks like:

```
[root@rancher-dev rancher]# system-docker pull alpine
Using default tag: latest
DEBU[0201] Calling GET /v1.23/info
> WARN[0201] Could not get operating system name: Error opening /usr/lib/os-release: open /usr/lib/os-release: no such file or directory
WARN[0201] Could not get operating system name: Error opening /usr/lib/os-release: open /usr/lib/os-release: no such file or directory
DEBU[0201] Calling POST /v1.23/images/create?fromImage=alpine%3Alatest
DEBU[0201] hostDir: /etc/docker/certs.d/10.10.10.23:5555
DEBU[0201] Trying to pull alpine from http://10.10.10.23:5555/ v2
DEBU[0204] Pulling ref from V2 registry: alpine:latest
DEBU[0204] pulling blob "sha256:2aecc7e1714b6fad58d13aedb0639011b37b86f743ba7b6a52d82bd03014b78e" latest: Pulling from library/alpine
DEBU[0204] Downloaded 2aecc7e1714b to tempfile /var/lib/system-docker/tmp/GetImageBlob281102233 2aecc7e1714b: Extracting  1.99 MB/1.99 MB
DEBU[0204] Untar time: 0.161064213s
DEBU[0204] Applied tar sha256:3fb66f713c9fa9debcdaa58bb9858bd04c17350d9614b7a250ec0ee527319e59 to 841c99a5995007d7a66b922be9bafdd38f8090af17295b4a44436ef433a2aecc7e1714b: Pull complete
Digest: sha256:0b94d1d1b5eb130dd0253374552445b39470653fb1a1ec2d81490948876e462c
Status: Downloaded newer image for alpine:latest
```
