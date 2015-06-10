---
title: Booting RancherOS with iPXE
layout: default

---
## Booting RancherOS via iPXE
----

```
#!ipxe
# Boot a persistent RangerOS to RAM

# Location of Kernel/Initrd images
set base-url http://releases.rancher.com/os/latest

kernel ${base-url}/vmlinuz CLOUD_CONFIG=http://example.com/cloud-config DEVICE=/dev/sda
initrd ${base-url}/initrd
boot
```

And an example cloud config can be found [here](http://rancherio.github.io/os/docs/cloud-config/).
