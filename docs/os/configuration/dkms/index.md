---
title: DKMS
layout: os-default

---

## DKMS

DKMS is supported by running the DKMS scripts inside a container. To compile any kernel modules, you first need to [install the kernel headers]({{site.baseurl}}/os/configuration/kernel-modules-kernel-headers/). After kernel headers are enabled, they are installed in `/lib/modules/$(uname -r)/build`. To deploy containers that runs DKMS, you will need to ensure that you bind mount in `/usr/src` and `/lib/modules`.  

### Docker Example

```
# Installing Kernel Headers for Docker
$ sudo ros service enable kernel-headers
$ sudo ros service up kernel-headers
# Run a container in Docker and bind mount specific directories to run DKMS
$ docker run -it -v /usr/src:/usr/src -v /lib/modules:/lib/modules ubuntu:15.10 sh -c 'apt-get update && apt-get install -y sysdig-dkms'
```

### System Docker Example

```
# Installing Kernel Headers for System Docker
$ sudo ros service enable kernel-headers-system-docker
$ sudo ros service up kernel-headers-system-docker
# Run a container in System Docker and bind mount specific directories to run DKMS
$ sudo system-docker run -it -v /usr/src:/usr/src -v /lib/modules:/lib/modules ubuntu:15.10 sh -c 'apt-get update && apt-get install -y sysdig-dkms'
```
