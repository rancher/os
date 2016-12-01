---
title: Rancher RancherOS in Openstack
layout: os-default
---

## Openstack    
---

As of v0.5.0, RancherOS releases include an Openstack image that can be found on our [releases page](https://github.com/rancher/os/releases). The image format is QCOW2. 

When launching an instance using the image, you must enable **Advanced Options** -> **Configuration Drive** and in order to use a [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) file.
