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


### cloud-init Datasources

Valid [datasources](https://github.com/rancher/os/blob/3338c4ac63597940bcde7e6005f1cc09287062a2/cmd/cloudinit/cloudinit.go#L378) for RancherOS.

| type | default |  
|---|---|
| ec2 | DefaultAddress | 
| file | path |
| url | url |
| cmdline |  |
| configdrive |  |
| digitalocean | DefaultAddress |
| gce |  |

### Cloud-Config
 
When booting via iPXE, RancherOS can be configured using a [cloud-config file]({{site.baseurl}}/os/configuration/#cloud-config).
