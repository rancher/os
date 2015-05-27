---
title: Cloud Config
layout: default

---

## Cloud Config through Cloud-Init
---

We currently support a very small portion of cloud-init. The cloud-init process is able to consume and execute data from user-data. Depending on the format of the information found, the process will behave differenty. If the user-data is a script (starting with the proper #!<interpreter>), we will execute it. If the user-data starts with `#cloud-config`, it will be processed by cloud-init. This cloud config is used for initial configuration on the very first boot of a server.  Without the `#cloud-config` as the first line, the cloud-init will not interpret the file as a cloud-config file.

A cloud config file uses a YAML format. YAML is easy to understand and easy to parse. For more information on YAML, please go [here](http://www.yaml.org/start.html). The most important formatting principle is indentation or whitespace. This indentation indicates relationships of the items to one another. If something is indented more than the previous line, it is a sub-item of the top item that is less indented.

Example: Notice how both are indented underneath `ssh-authorized-keys`.

```yaml
#cloud-config
ssh_authorized_keys:
  - ssh-rsa AAA...ZZZ example1@rancher
  - ssh-rsa BBB...ZZZ example1@rancher
```

In our example above, we have our `#cloud-config` line to indicate it's a cloud config file. We have 1 top-level key, `ssh_authorized_keys`. The values of the keys are the indented lines after the key.

### How does Cloud Config work in RancherOS

In RancherOS, we start with a default configuration. The cloud config file processed by cloud-init will extend and overwrite the default config. Finally, there is a `rancher.yml` file that will extend and overwrite the configuration. If you want to edit the `rancher.yml` file, please go [here]({{site.baseurl}}/docs/rancher-yml).

Typically, when you first boot the server, you'd pass in the cloud config file to configure the initialization of the server. After the first boot, if you have any changes for the configuration, it's recommended that you use `ros config` commands to set the `rancher` key in the configuration. Any changes will be saved in the `rancher.yml` file.

### Supported Cloud Init Directives

Please review the directives that we currently support in RancherOS.

```yaml
#cloud-config

# Adds SSH keys to the rancher user
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

# Anything you want to add to the rancher.yml must start with the rancher key
rancher:
 network:
  dns:
   nameservers
    - 8.8.8.8
    - 8.8.4.4

```
