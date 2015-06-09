---
title: Cloud Config
layout: default

---

## Configure RancherOS through Cloud Config
---

Cloud config is a declarative configuration file supported by many Linux distributions. A Linux OS supporting cloud config will invoke a `cloud-init` process during startup to parse the cloud config file and configure the operating system.

RancherOS runs its own `cloud-init` process in a system container. The `cloud-init` process will attempt to retrieve the
cloud config file from a variety of data sources. Once `cloud-init` obtains a cloud config file, it configures the Linux OS according to the content of the cloud config file.

When you create a RancherOS instance on AWS, for example, you can optionally specify a cloud config file. The cloud config file is then passed to the RancherOS instance as `user-data`. Inside the RanchreOS instance, the `cloud-init` process will retrieve the content of the cloud config file through the AWS cloud config data source: which simply extracts the content of `user-data` received by the VM instance. If the file starts with "`#cloud-config`", `cloud-init` will interpret that file as a cloud config file. If the file starts with `#!<interpreter>` (e.g., `#!/bin/sh`), `cloud-init` will simply execute that file. You can place any configuration commands in the file as scripts.

A cloud config file uses a YAML format. YAML is easy to understand and easy to parse. For more information on YAML, please go [here](http://www.yaml.org/start.html). The most important formatting principle is indentation or whitespace. This indentation indicates relationships of the items to one another. If something is indented more than the previous line, it is a sub-item of the top item that is less indented.

Example: Notice how both are indented underneath `ssh-authorized-keys`.

```yaml
#cloud-config
ssh_authorized_keys:
  - ssh-rsa AAA...ZZZ example1@rancher
  - ssh-rsa BBB...ZZZ example1@rancher
```

In our example above, we have our `#cloud-config` line to indicate it's a cloud config file. We have 1 top-level key, `ssh_authorized_keys`. The values of the keys are the indented lines after the key.

### How RancherOS Applies Cloud Config

RancherOS comes with a default configuration. The cloud config file processed by `cloud-init` will extend and overwrite the default configuration. Finally, the `rancher.yml` file will extend and overwrite the result of cloud config. You should not edit `rancher.yml` file directly. The `ros config` command allows you to change the content of the `rancher.yml` file.

Typically, when you first boot the server, you'd pass in the cloud config file to configure the initialization of the server. After the first boot, if you have any changes for the configuration, it's recommended that you use `ros config` commands to set the `rancher` key in the configuration. Any changes will be saved in the `rancher.yml` file.

### Supported Cloud Config Directives

RancherOS currently supports a small number of cloud config directives.

#### SSH Keys

You can add SSH keys to the default `rancher` user.

```yaml
# Adds SSH keys to the rancher user
ssh_authorized_keys:
  - ssh-rsa AAA... darren@rancher
```
#### Write Files to Disk

You can write files to the disk using the `write_files` directive.

```yaml
write_files:
  - path: /opt/rancher/bin/start.sh
    permissions: 0755
    owner: root
    content: |
      #!/bin/bash
      echo "I'm doing things on start"
```

#### Network Configuration

Network configuration section must start with the `rancher` key.

```yaml
rancher:
  network:
    interfaces:
      eth0:
        dhcp: false
        address: 192.168.100.100/24
        gateway: 192.168.100.1
        mtu: 1500
      eth1:
        dhcp: true
    dns:
      nameservers:
        - 8.8.8.8
        - 8.8.4.4

```
