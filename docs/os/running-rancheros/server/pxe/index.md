---
title: Booting RancherOS with iPXE
layout: os-default

---
## Booting RancherOS via iPXE
----

```
#!ipxe
# Boot a persistent RancherOS to RAM

# Location of Kernel/Initrd images
set base-url http://releases.rancher.com/os/latest

kernel ${base-url}/vmlinuz rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda] rancher.cloud_init.datasources=[url:http://example.com/cloud-config]
initrd ${base-url}/initrd
boot
```

### Hiding sensitive kernel commandline parameters

From RancherOS v0.9.0, secrets can be put on the `kernel` parameters line afer a `--` double dash, and they will be not be shown in any `/proc/cmdline`. These parameters
will be passed to the RancherOS init process and stored in the `root` accessible `/var/lib/rancher/conf/cloud-init.d/init.yml` file, and are available to the root user from the `ros config` commands.

For example, the `kernel` line above could be written as:

```
kernel ${base-url}/vmlinuz rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda] -- rancher.cloud_init.datasources=[url:http://example.com/cloud-config]
```

The hidden part of the command line can be accessed with either `sudo ros config get rancher.environment.EXTRA_CMDLINE`, or by using a service file's environment array.

An example service.yml file:

```
test:
  image: alpine
  command: echo "tell me a secret ${EXTRA_CMDLINE}"
  labels:
    io.rancher.os.scope: system
  environment:
  - EXTRA_CMDLINE
```

When this service is run, the `EXTRA_CMDLINE` will be set.


### cloud-init Datasources

Valid cloud-init datasources for RancherOS.

| type | default |  |
|---|---|--|
| ec2 | ec2's DefaultAddress |  |
| file | path |  |
| cmdline | /media/config-2 |  |
| configdrive |  |  |
| digitalocean | DefaultAddress |  |
| ec2 | DefaultAddress |  |
| file | path |  |
| gce |  |  |
| packet | DefaultAddress |  |
| url | url |  |
| vmware |  | set `guestinfo.cloud-init.config.data`, `guestinfo.cloud-init.config.data.encoding`, or `guestinfo.cloud-init.config.url` |
| * | This will add ["configdrive", "vmware", "ec2", "digitalocean", "packet", "gce"] into the list of datasources to try |  |

### Cloud-Config

When booting via iPXE, RancherOS can be configured using a [cloud-config file]({{site.baseurl}}/os/configuration/#cloud-config).

### VMware guestinfo

| GUESTINFO VARIABLE |	TYPE |
|---|---|
| `hostname |	hostname |
| `interface.<n>.name` |	string |
| `interface.<n>.mac` |	MAC address (is used to match the ethernet device's MAC address, not to set it) |
| `interface.<n>.dhcp` |	{"yes", "no"} |
| `interface.<n>.role` |	{"public", "private"} |
| `interface.<n>.ip.<m>.address` |	CIDR IP address |
| `interface.<n>.route.<l>.gateway` |	IP address |
| `interface.<n>.route.<l>.destination` |	CIDR IP address (not available yet) |
| `dns.server.<x>` | IP address |
| `dns.domain.<y> |	DNS search domain` |
| `cloud-init.config.data | string` |
| `cloud-init.config.data.encoding` |	{"", "base64", "gzip+base64"} |
| `cloud-init.config.url` |	URL |

> **Note:** "n", "m", "l", "x" and "y" are 0-indexed, incrementing integers. The identifier for an interface (`<n>`) is used in the generation of the default interface name in the form `eth<n>`.
