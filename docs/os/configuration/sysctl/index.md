---
title: sysctl Settings in RancherOS
layout: os-default

---

## sysctl Settings
---

The `rancher.sysctl` cloud-config key can be used to control sysctl parameters. This works in a manner similar to `/etc/sysctl.conf` for other Linux distros.

```
#cloud-config
rancher:
  sysctl:
    net.ipv4.conf.default.rp_filter: 1
```
