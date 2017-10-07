---
title: Rancher RancherOS in Openstack
layout: os-default
---

## Openstack
---

As of v0.5.0, RancherOS releases include an Openstack image that can be found on our [releases page](https://github.com/rancher/os/releases). The image format is QCOW2.

When launching an instance using the image, you must enable ConfigrDrive or Metadata Agent, and in order to use a [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file.
RancherOS fetch user data from Metadata Agent first. If can't fetch from Metadata Agent, then fetch from ConfigDrive.
If you use ConfigDrive, you must enable **Advanced Options** -> **Configuration Drive** .
