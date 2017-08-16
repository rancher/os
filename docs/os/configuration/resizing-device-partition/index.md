---
title: Resizing a Device Partition in RancherOS


---

## Resizing a Device Partition
---

The `resize_device` cloud config option can be used to automatically extend the first partition (assuming its `ext4`) to fill the size of it's device.

Once the partition has been resized to fill the device, a `/var/lib/rancher/resizefs.done` file will be written to prevent the resize tools from being run again. If you need it to run again, delete that file and reboot.

```yaml
#cloud-config
rancher:
  resize_device: /dev/sda
```

This behavior is the default when launching RancherOS on AWS.
