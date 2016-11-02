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

### Datasources 

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
