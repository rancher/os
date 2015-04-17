---
title: Configuring Cloud Init
layout: default

---

## Supported Cloud Init Functionality
---
We currently support a very small portion of cloud-init.  If the user_data is a script (starting with the proper #!<interpreter>), we will execute it.  If the user_data starts with `#cloud-config` it will be processed by cloud-init.  The below directives are supported.  Using the `rancher` key you can also configure anything found in [rancher.yml]({{site.baseurl}}/docs/configuring/).

```yaml
#cloud-config

ssh_authorized_keys:
 - ssh-rsa AAA... darren@rancher

write_files:
 write_files:
  - path: /opt/rancher/bin/start.sh
    permissions: 0755
    owner: root
    content: |
     #!/bin/bash
     echo "I'm doing things on start"

# Anything you can put in the rancher.yml
rancher:
 network:
  dns:
   nameservers
    - 8.8.8.8
    - 8.8.4.4

```

