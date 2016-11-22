---
title: System Services in RancherOS
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
$ sudo ros service up kernel-headers
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


