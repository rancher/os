---
title: Resizing a Device Partition in RancherOS
layout: os-default

---

## Resizing a Device Partition
---

The `resize_device` cloud config option can be used to automatically extend the first partition to fill the size of it's device.

```yaml
#cloud-config
rancher:
  resize_device: /dev/sda
```

This behavior is the default when launching RancherOS on AWS.
