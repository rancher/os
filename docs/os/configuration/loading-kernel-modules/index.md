---
title: Loading Kernel Modules in RancherOS
layout: os-default

---

## Loading Kernel Modules
---

Kernel modules can be automatically loaded with the `rancher.modules` cloud-config field.

```yaml
#cloud-config
rancher:
  modules: [btrfs]
```

This functionality is also available via a kernel parameter. For example, the btrfs module could be automatically loaded with `rancher.modules=[btrfs]` as a kernel parameter.

### Loading Kernel Modules via a System Service

Privileged containers can load kernel modules. In RancherOS, the kernel modules are in the standard `/lib/modules/$(uname -r)` folder.  If you want to be able to run `modprobe` from a container, you will need to bind mount the `/lib/modules` into your container.   

```yaml
myservice:
  image: ...
  privileged: true
  volumes:
  - /lib/modules:/lib/modules
```

By default, the `/lib/modules` folder is already available in the console.
