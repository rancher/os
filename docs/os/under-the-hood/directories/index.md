---
title: Directories
layout: os-default
---

## How directories are mounted, Ram disk, etc.
---

### Persistent Directories Across Reboot

With v0.4.0, ubuntu and debian consoles are [persistent consoles]({{site.baseurl}}/os/configuration/custom-console/#console-persistence). Therefore, the only difference is what is persisted inside a containers as opposed to on the host. If a container is deleted/rebuilt, state in the console will be lost except what is in the persisted directories.

```
/home
/opt
/var/lib/docker
/var/lib/rancher
```

