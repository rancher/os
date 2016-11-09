---
title: Configuring RancherOS
layout: os-default
redirect_from:
  - os/cloud-config/
---

## Configuring RancherOS
---
There are two ways that RancherOS can be configured.

1. A cloud-config file can be used to provide configuration when first booting RancherOS.

2. Manually changing configuration with the `ros config` command.

Typically, when you first boot the server, you pass in a cloud-config file to configure the initialization of the server. After the first boot, if you have any changes for the configuration, it's recommended that you use `ros config` to set the necessary configuration properties. Any changes will be saved on disk and a reboot will be required for changes to be applied.

### Cloud-Config

Cloud-config is a declarative configuration file format supported by many Linux distributions and is the primary configuration mechanism for RancherOS.

A Linux OS supporting cloud-config will invoke a cloud-init process during startup to parse the cloud-config file and configure the operating system. RancherOS runs its own cloud-init process in a system container. The cloud-init process will attempt to retrieve a cloud-config file from a variety of data sources. Once cloud-init obtains a cloud-config file, it configures the Linux OS according to the content of the cloud-config file.

When you create a RancherOS instance on AWS, for example, you can optionally provide cloud-config passed in the `user-data` field. Inside the RancherOS instance, cloud-init process will retrieve the cloud-config content through its AWS cloud-config data source, which simply extracts the content of user-data received by the VM instance. If the file starts with "`#cloud-config`", cloud-init will interpret that file as a cloud-config file. If the file starts with `#!<interpreter>` (e.g., `#!/bin/sh`), cloud-init will simply execute that file. You can place any configuration commands in the file as scripts.

A cloud-config file uses the YAML format. YAML is easy to understand and easy to parse. For more information on YAML, please read more at the [YAML site](http://www.yaml.org/start.html). The most important formatting principle is indentation or whitespace. This indentation indicates relationships of the items to one another. If something is indented more than the previous line, it is a sub-item of the top item that is less indented.

Example: Notice how both are indented underneath `ssh-authorized-keys`.

```yaml
#cloud-config
ssh_authorized_keys:
  - ssh-rsa AAA...ZZZ example1@rancher
  - ssh-rsa BBB...ZZZ example2@rancher
```

In our example above, we have our `#cloud-config` line to indicate it's a cloud-config file. We have 1 top-level property, `ssh_authorized_keys`. Its value is a list of public keys that are represented as a dashed list under `ssh_authorized_keys:`.

### Manually Changing Configuration

To update RancherOS configuration after booting, the `ros config` command can be used.

#### Getting Values

You can easily get any value that's been set in the `/var/lib/rancher/conf/cloud-config.yml` file. Let's see how easy it is to get the DNS configuration of the system.

```
$ sudo ros config get rancher.network.dns.nameservers
- 8.8.8.8
- 8.8.4.4
```

#### Setting Values 

You can set values in the `/var/lib/rancher/conf/cloud-config.yml` file.

Setting a simple value in the `/var/lib/rancher/conf/cloud-config.yml`

```
$ sudo ros config set rancher.docker.tls true
```

Setting a list in the `/var/lib/rancher/conf/cloud-config.yml`

```
$ sudo ros config set rancher.network.dns.nameservers "['8.8.8.8','8.8.4.4']"
```

#### Exporting the Current Configuration

To output and review the current configuration state you can use the `ros config export` command.

```
$ sudo ros config export
rancher:
  docker:
    tls: true
  network:
    dns:
      nameservers:
      - 8.8.8.8
      - 8.8.4.4
```

#### Validating a Configuration File

To validate a configuration file you can use the `ros config validate` command.

```
$ sudo ros config validate -i cloud-config.yml
```
