---
title: Booting RancherOS with iPXE
layout: default

---
## Booting RancherOS via iPXE
----

```
#!ipxe
# Boot a persistent RancherOS to RAM

# Location of Kernel/Initrd images
set base-url http://releases.rancher.com/os/latest

kernel ${base-url}/vmlinuz rancher.state.autoformat=[/dev/sda] rancher.cloud_init.datasources=[url:http://example.com/cloud-config]
initrd ${base-url}/initrd
boot
```

Valid [datasources](https://github.com/rancherio/os/blob/3338c4ac63597940bcde7e6005f1cc09287062a2/cmd/cloudinit/cloudinit.go#L378) are:

| type | default |  
|---|---|
| ec2 | DefaultAddress | 
| file | path |
| url | url |
| cmdline |  |
| configdrive |  |
| digitalocean | DefaultAddress |
| gce |  |

And an example cloud config can be found [here](http://rancherio.github.io/os/docs/cloud-config/).
