---
title: Configuring Cloud Init
layout: default

---

## Cloud Init

```yaml
cloud_init:
datasources:
- configdrive:/media/config-2
```

In the rancher.yml you can configure which datasources to use for cloud-init.  Multiple datasources can be set but the datasource that is available the fastest will be used.  This value is usually prepopulated with the current setting for your environment.  Valid value are:

1. `configdrive:PATH` - Look for an OpenStack compatible config drive mounted at `PATH`
1. `file:PATH` - Read the `FILE` as the user data.
1. `ec2` - Look for EC2 style meta data at 169.254.169.254
1. `ec2:IP_ADDRESS` - Look for EC2 style meta data at the `IP_ADDRESS`
1. `url:URL` - Download `URL` and use that as the user data
1. `cmdline:URL` - Look for `cloud-config-url=URL` in `/proc/cmdline` and download `URL` as user data

## Cloud Init

We currently support a very small portion of cloud-init.  If the user_data is a script (starting with the proper #!<interpreter>) we will execute it.  If the user_data starts with `#cloud-config` it will be processed by cloud-init.  The below directives are supported.  Using the `rancher` key you can also configure anything found in [`rancher.yml`](docs/config.md).

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