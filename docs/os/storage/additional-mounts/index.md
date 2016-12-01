---
title: Additional Mounts in RancherOS
layout: os-default
redirect_from:
  - os/configuration/additional-mounts/
---

## Additional Mounts

Additional mounts can be specified as part of your [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config). These mounts are applied within the console container. Here's a simple example that mounts `/dev/vdb` to `/mnt/s`.

```yaml
#cloud-config
mounts:
- ["/dev/vdb", "/mnt/s", "ext4", ""]
```

<br>

The four arguments for each mount are the same as those given for [cloud-init](https://cloudinit.readthedocs.io/en/latest/topics/examples.html#adjust-mount-points-mounted). Only the first four arguments are currently supported. The `mount_default_fields` key is not yet implemented.

RancherOS uses the mount syscall rather than the `mount` command behind the scenes. This means that `auto` cannot be used as the filesystem type (third argument) and `defaults` cannot be used for the options (forth argument).

### Shared Mounts

By default, `/media` and `/mnt` are mounted as shared in the console container. This means that mounts within these directories will propogate to the host as well as other system services that mount these folders as shared.

See [here](https://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt) for a more detailed overview of shared mounts and their properties.
