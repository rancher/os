---
title: System Docker Volumes
layout: os-default

---

## System Docker Volumes
---

A few services are containers in `created` state. Their purpose is to provide volumes for other services.

### user-volumes

Provides user accessible persistent storage directories, used by console service:

```
/home
/opt
```

### container-data-volumes 

Provides docker storage directory, used by console service (and, indirectly, by docker)

```
/var/lib/docker
```

### command-volumes 

Provides necessary command binaries (read-only), used by system services:

```
/usr/bin/docker-containerd.dist
/usr/bin/docker-containerd-shim.dist
/usr/bin/docker-runc.dist
/usr/bin/docker.dist
/usr/bin/dockerlaunch
/usr/bin/user-docker
/usr/bin/system-docker
/sbin/poweroff
/sbin/reboot
/sbin/halt
/sbin/shutdown
/usr/bin/respawn
/usr/bin/ros
/usr/bin/cloud-init
/usr/sbin/netconf
/usr/sbin/wait-for-docker
/usr/bin/switch-console
```

### system-volumes

Provides necessary persistent directories, used by system services:

```
/host/dev
/etc/docker
/etc/hosts
/etc/resolv.conf
/etc/ssl/certs/ca-certificates.crt.rancher
/etc/selinux
/lib/firmware
/lib/modules
/run
/usr/share/ros
/var/lib/rancher/cache
/var/lib/rancher/conf
/var/lib/rancher
/var/log
/var/run
```

### all-volumes

Combines all of the above, used by the console service.


