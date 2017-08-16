---
title: Custom System Services in RancherOS

---

## Custom System Services

You can also create your own system service in [Docker Compose](https://docs.docker.com/compose/) format. After creating your own custom service, you can launch it in RancherOS in a couple of methods. The service could be directly added to the [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config), or a `docker-compose.yml` file could be saved at a http(s) url location or in a specific directory of RancherOS.

### Launching Services through Cloud-Config

If you want to boot RancherOS with a system service running, you can add the service to the cloud-config that is passed to RancherOS. When RancherOS starts, this service will automatically be started.

```yaml
#cloud-config
rancher:
  services:
    nginxapp:
      image: nginx
      restart: always
```

### Launching Services using local files

If you already have RancherOS running, you can start a system service by saving a `docker-compose.yml` file at `/var/lib/rancher/conf/`.

```yaml
nginxapp:
  image: nginx
  restart: always
```

To enable a custom system service from the file location, the command must indicate the file location if saved in RancherOS. If the file is saved at a http(s) url, just use the http(s) url when enabling/disabling.

```
# Enable the system service saved in /var/lib/rancher/conf
$ sudo ros service enable /var/lib/rancher/conf/example.yml
# Enable a system service saved at a http(s) url
$ sudo ros service enable https://mydomain.com/example.yml
```

<br>

After the custom system service is enabled, you can start the service using `sudo ros service up <serviceName>`. The `<serviceName>` will be the names of the services inside the `docker-compose.yml`.

```
$ sudo ros service up nginxapp
# If you have more than 1 service in your docker-compose.yml, add all service names to the command
$ sudo ros service up service1 service2 service3
```

### Launching Services from a web repository

The https://github.com/rancher/os-services repository is used for the built-in services, but you can create your own, and configure RancherOS to use it in addition (or to replace) it.

The config settings to set the url in which `ros` should look for an `index.yml` file is: `rancher.repositories.<name>.url`. The `core` repository url is set when a release is made, and any other `<name>` url you add will be listed together when running `ros console list`, `ros service list` or `ros engine list`

For example, in RancherOS v0.7.0, the `core` repository is set to `https://raw.githubusercontent.com/rancher/os-services/v0.7.0`.

### Service development and testing

If you're building your own services in a branch on GitHub, you can push to it, and then load your service from there.

For example, when developing the zfs service:

```
rancher@zfs:~$ sudo ros config set rancher.repositories.zfs.url https://raw.githubusercontent.com/SvenDowideit/os-services/zfs-service
rancher@zfs:~$ sudo ros service list
disabled amazon-ecs-agent
disabled kernel-extras
enabled  kernel-headers
disabled kernel-headers-system-docker
disabled open-vm-tools
disabled amazon-ecs-agent
disabled kernel-extras
disabled kernel-headers
disabled kernel-headers-system-docker
disabled open-vm-tools
disabled zfs
[rancher@zfs ~]$ sudo ros service enable zfs
Pulling zfs (zombie/zfs)...
latest: Pulling from zombie/zfs
b3e1c725a85f: Pull complete
4daad8bdde31: Pull complete
63fe8c0068a8: Pull complete
4a70713c436f: Pull complete
bd842a2105a8: Pull complete
d1a8c0826fbb: Pull complete
5f1c5ffdf34c: Pull complete
66c2263f2388: Pull complete
Digest: sha256:eab7b8c21fbefb55f7ee311dd236acee215cb6a5d22942844178b8c6d4e02cd9
Status: Downloaded newer image for zombie/zfs:latest
[rancher@zfs ~]$ sudo ros service up zfs
WARN[0000] The KERNEL_VERSION variable is not set. Substituting a blank string.
INFO[0000] Project [os]: Starting project
INFO[0000] [0/21] [zfs]: Starting
INFO[0000] [1/21] [zfs]: Started
INFO[0000] Project [os]: Project started

```

Beware that there is an overly aggressive caching of yml files - so when you push a new yml file to your repo, you need to
delete the files in `/var/lib/rancher/cache`.

The image that you specify in the service yml file needs to be pullable - either from a private registry, or on the Docker Hub.

### Service cron

RancherOS has a system cron service based on [Container Crontab](https://github.com/rancher/container-crontab). This can be used to start, restart or stop system containers.

To use this on your service, add a `cron.schedule` label to your service's description:

```
my-service:
  image: namespace/my-service:v1.0.0
  command: my-command
  labels:
    io.rancher.os.scope: "system"
    cron.schedule: "0 * * * * ?"
```

For a cron service that can be used with user Docker containers, see the `crontab` system service.

### Service log rotation

RancherOS provides a built in `logrotate` container that makes use of logrotate(8) to rotate system logs. This is called on an hourly basis by the `system-cron` container.

If you would like to make use of system log rotation for your system service, do the following.

Add `system-volumes` to your service description's `volumes_from` section. You could also use a volume group containing `system-volumes` e.g. `all-volumes`.

```
my-service:
  image: namespace/my-service:v1.0.0
  command: my-command
  labels:
    io.rancher.os.scope: "system"
  volumes_from:
    - system-volumes
```

Next, add an entry point script to your image and copy your logrotate configs to `/etc/logrotate.d/` on startup.

Example Dockerfile:
```
FROM alpine:latest
COPY logrotate-myservice.conf entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
```

Example entrypoint.sh (Ensure that this script has the execute bit set).
```
#!/bin/sh

cp logrotate-myservice.conf /etc/logrotate.d/myservice

exec "$@"
```

Your service's log rotation config will now be included when the system logrotate runs. You can view logrotate output with `system-docker logs logrotate`.

### Creating your own Console

Once you have your own Services repository, you can add a new service to its index.yml, and then add a `<service-name>.yml` file to the directory starting with the first letter.

To create your own console images, you need to:

1 install some basic tools, including an ssh daemon, sudo, and kernel module tools
2 create `rancher` and `docker` users and groups with UID and GID's of `1100` and `1101` respectively
3 add both users to the `docker` and `sudo` groups
4 add both groups into the `/etc/sudoers` file to allow password-less sudo
5 configure sshd to accept logins from users in the `docker` group, and deny `root`.
6 set `ENTRYPOINT ["/usr/bin/ros", "entrypoint"]`

the `ros` binary, and other host specific configuration files will be bind mounted into the running console container when its launched.

For examples of existing images, see https://github.com/rancher/os-images.

## Labels

We use labels to determine how to handle the service containers.

Key | Value |Description
----|-----|---
`io.rancher.os.detach` | Default: `true` | Equivalent of `docker run -d`. If set to `false`, equivalent of `docker run --detach=false`
`io.rancher.os.scope` | `system` | Use this label to have the container deployed in System Docker instead of Docker.
`io.rancher.os.before`/`io.rancher.os.after` | Service Names (Comma separated list is accepted) | Used to determine order of when containers should be started.
`io.rancher.os.createonly` | Default: `false` | When set to `true`, only a `docker create` will be performed and not a `docker start`.
`io.rancher.os.reloadconfig` | Default: `false`| When set to `true`, it reloads the configuration.


RancherOS uses labels to determine if the container should be deployed in System Docker. By default without the label, the container will be deployed in User Docker.

```yaml
labels:
  - io.rancher.os.scope=system
```


### Example of how to order container deployment

```yaml
foo:
  labels:
    # Start foo before bar is launched
    io.rancher.os.before: bar
    # Start foo after baz has been launched
    io.rancher.os.after: baz
```
