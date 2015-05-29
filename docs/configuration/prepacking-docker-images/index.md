---
title: Pre-packing Docker Images
layout: default

---

## Pre-packing Docker Images
---

On boot RancherOS scans `/var/lib/docker/preload` and `/var/lib/system-docker/preload` dirs and tries to load container image archives it finds there, with `docker load` and `system-docker load`.   

The archives are `.tar` files, optionally compressed with `xz` or `gzip`. These can be produced by `docker save` command, e.g.:

```
docker save my-image1 my-image2 some-other/image3 | xz > my-images.tar.xz
```

The resulting files should be placed into `/var/lib/docker/preload` or `/var/lib/system-docker/preload` (depending on whether you want it preloaded into docker or system-docker).

Pre-loading process only reads each new archive once, so it won't take time on subsequent boots (`<archive>.done` files are created to mark the read archives). If you update the archive (place a newer archive with the same name) it'll get read on the next boot as well.
 
Pre-packing docker images is handy when you're customizing your RancherOS distribution (perhaps, building cloud VM images for your infrastructure). You might be interested in [os-installer](https://github.com/rancherio/os-installer) for this purpose.
