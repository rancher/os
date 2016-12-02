---
title: Persistent State Partition in RancherOS
layout: os-default
---

## Persistent State Partition

RancherOS will store its state in a single partition specified by the `dev` field.  The field can be a device such as `/dev/sda1` or a logical name such `LABEL=state` or `UUID=123124`.  The default value is `LABEL=RANCHER_STATE`.  The file system type of that partition can be set to `auto` or a specific file system type such as `ext4`.

```yaml
#cloud-config
rancher:
  state:
    fstype: auto
    dev: LABEL=RANCHER_STATE
    autoformat:
    - /dev/sda
    - /dev/vda
```

### Autoformat

You can specify a list of devices to check to format on boot.  If the state partition is already found, RancherOS will not try to auto format a partition. By default, auto-formatting is off.

RancherOS will autoformat the partition to ext4 if the device specified in `autoformat`:

* Contains a boot2docker magic string
* Starts with 1 megabyte of zeros and `rancher.state.formatzero` is true
