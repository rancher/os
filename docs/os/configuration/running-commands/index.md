---
title: Running Commands in RancherOS
layout: os-default

---

## Running Commands
---

You can automate running commands on boot using the `runcmd` cloud-config directive. Commands can be specified as either a list or a string. In the latter case, the command is executed with `sh`.

```yaml
#cloud-config
runcmd:
- [ touch, /home/rancher/test1 ]
- echo "test" > /home/rancher/test2
```

Commands specified using `runcmd` will be executed within the context of the `console` container. More details on the ordering of commands run in the `console` container can be found [here]({{site.baseurl}}/os/system-services/built-in-system-services/#console).

### Running Docker commands

When using `runcmd`, RancherOS will wait for all commands to complete before starting Docker. As a result, any `docker run` command should not be placed under `runcmd`. Instead, the `/etc/rc.local` script can be used. RancherOS will not wait for commands in this script to complete, so you can use the `wait-for-docker` command to ensure that the Docker daemon is running before performing any `docker run` commands.

```yaml
#cloud-config
rancher:
write_files:
  - path: /etc/rc.local
    permissions: "0755"
    owner: root
    content: |
      #!/bin/bash
      wait-for-docker
      docker run -d nginx
```

Running Docker commands in this manner is useful when pieces of the `docker run` command are dynamically generated. For services whose configuration is static, [adding a system service]({{site.baseurl}}/os/system-services/adding-system-services/) is recommended.

## Running Commands Early in the Boot Process
---

The `bootcmd` parameter can be used to run commands earlier in the boot process. In particular, `bootcmd` will be executed while RancherOS is still running from memory and before System Docker and any system services are started.

The syntax for bootcmd is the same as `runcmd`.

```yaml
#cloud-config
bootcmd:
- [ mdadm, --assemble, --scan ]
```
