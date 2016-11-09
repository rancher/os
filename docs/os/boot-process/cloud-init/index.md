---
title: Cloud-init
layout: os-default

---

## Cloud-init

Userdata and metadata can be fetched from a cloud provider, VM runtime, or management service during the RancherOS boot process. Since v0.8.0, this process occurs while RancherOS is still running from memory and before System Docker starts. It is configured by the `rancher.cloud_init.datasources` configuration parameter. For cloud-provider specific images, such as AWS and GCE, the datasource is pre-configured.

### Userdata

Userdata is a file given by users when launching RancherOS hosts. It is stored in different locations depending on its format. If the userdata is a [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file, indicated by beginning with `#cloud-config` and being in YAML format, it is stored in `/var/lib/rancher/conf/cloud-config.d/boot.yml`. If the userdata is a script, indicated by beginning with `#!`, it is stored in `/var/lib/rancher/conf/cloud-config-script`.

### Metadata

Although the specifics vary based on provider, a metadata file will typically contain information about the RancherOS host and contain additional configuration. Its primary purpose within RancherOS is to provide an alternate source for SSH keys and hostname configuration. For example, AWS launches hosts with a set of authorized keys and RancherOS obtains these via metadata. Metadata is stored in `/var/lib/rancher/conf/metadata`.

## Configuration Load Order

[Cloud-config]({{site.baseurl}}/os/configuration/#cloud-config/) is read by system services when they need to get configuration. Each additional file overwrites and extends the previous configuration file.

1. `/usr/share/ros/os-config.yml` - This is the system default configuration, which should **not** be modified by users.
2. `/usr/share/ros/oem/oem-config.yml` - This will typically exist by OEM, which should **not** be modified by users.
3. Files in `/var/lib/rancher/conf/cloud-config.d/` ordered by filename. If a file is passed in through user-data, it is written by cloud-init and saved as `/var/lib/rancher/conf/cloud-config.d/boot.yml`.
4. `/var/lib/rancher/conf/cloud-config.yml` - If you set anything with `ros config set`, the changes are saved in this file.
5. Kernel parameters with names starting with `rancher`.
6. `/var/lib/rancher/conf/metadata` - Metadata added by cloud-init.
