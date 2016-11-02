---
title: Configuring DNS in RancherOS
layout: os-default
redirect_from:
  - os/configuration/networking/#dns
---

## DNS

If you wanted to configure the DNS through the cloud config file, you'll need to place DNS configurations within the `rancher` key.

```yaml
#cloud-config

#Remember, any changes for rancher will be within the rancher key
rancher:
  network:
    dns:
      search:
        - mydomain.com
        - example.com
```

Using `ros config`, you can set the `nameservers`, and `search`, which directly map to the fields of the same name in `/etc/resolv.conf`.

```
$ sudo ros config set rancher.network.dns.search "['mydomain.com','example.com']"
$ sudo ros config get rancher.network.dns.search
- mydomain.com
- example.com
```
