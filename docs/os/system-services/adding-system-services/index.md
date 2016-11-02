---
title: Adding System Services in RancherOS
layout: os-default
redirect_from:
  - os/system-services/
  - os/configuration/system-services/
---

## System Services

A system service is a container that can be run in either System Docker or Docker. Rancher provides services that are already available in RancherOS by adding them to the [os-services repo](https://github.com/rancher/os-services). Anything in the `index.yml` file from the repository for the tagged release will be an available system service when using the `ros service list` command.

### Enabling and Starting System Services

For any services that are listed from the `ros service list`, they can be enabled by running a single command. After enabling a service, you will need to run start the service.

```
# List out available system services
$ sudo ros service list
disabled amazon-ecs-agent
disabled kernel-headers
disabled kernel-headers-system-docker
disabled open-vm-tools
# Enable a system service
$ sudo ros service enable kernel-headers
# Start a system service
$ sudo ros service up -d kernel-headers
```

### Disabling and Removing System Services

In order to stop a system service from running, you will need to stop and disable the system service.

```
# List out available system services
$ sudo ros service list
disabled amazon-ecs-agent
enabled kernel-headers
disabled kernel-headers-system-docker
disabled open-vm-tools
# Disable a system service
$ sudo ros service disable kernel-headers
# Stop a system service
$ sudo ros service stop kernel-headers
# Remove the containers associated with the system service
$ sudo ros service down kernel-headers
```

<br>
If you want to remove a system service from the list of service, just delete the service.

```
$ sudo ros service delete <serviceName>
```

### Custom System Services

You can also create your own system service in [Docker Compose](https://docs.docker.com/compose/) format. After creating your own custom service, you can launch it in RancherOS in a couple of methods. The service could be directly added to the [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config), or a `docker-compose.yml` file could be saved at a http(s) url location or in a specific directory of RancherOS.

#### Launching Services through Cloud-Config

If you want to boot RancherOS with a system service running, you can add the service to the cloud-config that is passed to RancherOS. When RancherOS starts, this service will automatically be started.

```yaml
#cloud-config
rancher:
  services:
    nginxapp:
      image: nginx
      restart: always
```      

#### Launching Custom System Services inside RancherOS

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

After the custom system service is enabled, you can start the service using `sudo ros service up -d <serviceName>`. The `<serviceName>` will be the names of the services inside the `docker-compose.yml`.

```
$ sudo ros service up -d nginxapp
# If you have more than 1 service in your docker-compose.yml, add all service names to the command
$ sudo ros service up -d service1 service2 service3
```

### System Docker vs. Docker

RancherOS uses labels to determine if the container should be deployed in System Docker. By default without the label, the container will be deployed in Docker.

```yaml
labels:
  - io.rancher.os.scope=system
```

### Labels

We use labels to determine how to handle the service containers.

Key | Value |Description
----|-----|---
`io.rancher.os.detach` | Default: `true` | Equivalent of `docker run -d`. If set to `false`, equivalent of `docker run --detach=false`
`io.rancher.os.scope` | `system` | Use this label to have the container deployed in System Docker instead of Docker.
`io.rancher.os.before`/`io.rancher.os.after` | Service Names (Comma separated list is accepted) | Used to determine order of when containers should be started.
`io.rancher.os.createonly` | Default: `false` | When set to `true`, only a `docker create` will be performed and not a `docker start`.
`io.rancher.os.reloadconfig` | Default: `false`| When set to `true`, it reloads the configuration.


#### Example of how to order container deployment

```yaml
foo:
  labels:
    # Start foo before bar is launched
    io.rancher.os.before: bar
    # Start foo after baz has been launched
    io.rancher.os.after: baz
```
