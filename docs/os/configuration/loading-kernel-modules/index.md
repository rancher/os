---
title: Loading Kernel Modules in RancherOS
layout: os-default

---

## Loading Kernel Modules
---

Privileged containers can load kernel modules. In RancherOS, the kernel modules are in the standard `/lib/modules/$(uname -r)` folder.  If you want to be able to run `modprobe` from a container, you will need to bind mount the `/lib/modules` into your container.   

```yaml
myservice:
  image: ...
  privileged: true
  volumes:
  - /lib/modules:/lib/modules
```

By default, the `/lib/modules` folder is already available in the console.
