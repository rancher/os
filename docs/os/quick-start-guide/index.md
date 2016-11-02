---
title: Quick Start Guide for RancherOS
layout: os-default

---

## Quick Start Guide
---

If you have a specific RanchersOS machine requirements, please check out our [guides on running RancherOS]({{site.baseurl}}/os/running-rancheros/). With the rest of this guide, we'll start up a RancherOS using [Docker machine]({{site.baseurl}}/os/running-rancheros/workstation/docker-machine/) and show you some of what RancherOS can do.

### Launching RancherOS using Docker Machine

Before moving forward, you'll need to have [Docker Machine](https://docs.docker.com/machine/) and [VirtualBox](https://www.virtualbox.org/wiki/Downloads) installed. Once you have VirtualBox and Docker Machine installed, it's just one command to get RancherOS running. 

```
$ docker-machine create -d virtualbox --virtualbox-boot2docker-url https://releases.rancher.com/os/latest/rancheros.iso <MACHINE-NAME>
```

That's it! You're up and running a RancherOS instance.

To log into the instance, just use the `docker-machine` command.

```
$ docker-machine ssh <MACHINE-NAME>
```

### A First Look At RancherOS

There are two Docker daemons running in RancherOS. The first is called **System Docker**, which is where RancherOS runs system services like ntpd and syslog. You can use the `system-docker` command to control the **System Docker** daemon. 

The other Docker daemon running on the system is **Docker**, which can be accessed by using the normal `docker` command.

When you first launch RancherOS, there are no containers running in the Docker daemon. However, if you run the same command against the System Docker, you’ll see a number of system services that are shipped with RancherOS. 

> **Note:** `system-docker` can only be used by root, so it is necessary to use the `sudo` command whenever you want to interact with System Docker.

```
$ sudo system-docker ps
CONTAINER ID        IMAGE                       COMMAND                  CREATED             STATUS              PORTS               NAMES
6f56057cf5ba        rancher/os-base:v0.5.0      "/usr/sbin/entry.sh /"   16 seconds ago      Up 15 seconds                           docker
bd5376830237        rancher/os-console:v0.5.0   "/usr/sbin/entry.sh /"   16 seconds ago      Up 15 seconds                           console
ede8ce39fff5        rancher/os-base:v0.5.0      "/usr/sbin/entry.sh n"   16 seconds ago      Up 15 seconds                           network
9e5d18bca391        rancher/os-base:v0.5.0      "/usr/sbin/entry.sh n"   17 seconds ago      Up 16 seconds                           ntp
393b9fb7e30a        rancher/os-udev:v0.5.0      "/usr/sbin/entry.sh /"   18 seconds ago      Up 16 seconds                           udev
dc2cafca3c69        rancher/os-syslog:v0.5.0    "/usr/sbin/entry.sh /"   18 seconds ago      Up 17 seconds                           syslog
439d5535fbfa        rancher/os-base:v0.5.0      "/usr/sbin/entry.sh /"   18 seconds ago      Up 17 seconds                           acpid
```

Some containers are run at boot time, and others, such as the `console`, `docker`, etc. containers are always running.

## Using RancherOS
---

### Deploying a Docker Container

Let's try to deploy a normal Docker container on the Docker daemon.  The RancherOS Docker daemon is identical to any other Docker environment, so all normal Docker commands work.

```
$ docker run -d nginx 
```

You can see that the nginx container is up and running:

```
$ docker ps
CONTAINER ID        IMAGE               COMMAND                  CREATED             STATUS              PORTS               NAMES
e99c2c4b8b30        nginx               "nginx -g 'daemon off"   12 seconds ago      Up 11 seconds       80/tcp, 443/tcp     drunk_ptolemy
```

### Deploying A System Service Container

The following is a simple Docker container to set up Linux-dash, which is a minimal low-overhead web dashboard for monitoring Linux servers. The Dockerfile will be like this:

```
FROM hwestphal/nodebox
MAINTAINER hussein.galal.ahmed.11@gmail.com

RUN opkg-install unzip
RUN curl -k -L -o master.zip https://github.com/afaqurk/linux-dash/archive/master.zip
RUN unzip master.zip
WORKDIR linux-dash-master
RUN npm install

ENTRYPOINT ["node","server"]
```

Using the `hwestphal/nodebox` image, which uses a Busybox image and installs `node.js` and `npm`. We downloaded the source code of Linux-dash, and then ran the server. Linux-dash will run on port 80 by default.

To run this container in System Docker use the following command:

```
$ sudo system-docker run -d --net=host --name busydash husseingalal/busydash
```
In the commad, we used `--net=host` to tell System Docker not to containerize the container's networking, and use the host’s networking instead. After running the container, you can see the monitoring server by accessing `http://<IP_OF_MACHINE`.

![System Docker Container]({{site.baseurl}}/img/os/Rancher_busydash.png)

To make the container survive during the reboots, you can create the `/opt/rancher/bin/start.sh` script, and add the Docker start line to launch the Docker at each startup.

```
$ sudo mkdir -p /opt/rancher/bin
$ echo “sudo system-docker start busydash” | sudo tee -a /opt/rancher/bin/start.sh
$ sudo chmod 755 /opt/rancher/bin/start.sh
```

### Using ROS

Another useful command that can be used with RancherOS is `ros` which can be used to control and configure the system. 

```
$ ros -v
ros version 0.0.1
```

RancherOS state is controlled by a cloud config file. `ros` is used to edit the configuration of the system, to see for example the dns configuration of the system:

```
$ sudo ros config get rancher.dns
- 8.8.8.8
- 8.8.4.4
```


When using the native Busybox console, any changes to the console will be lost after reboots, only changes to `/home` or `/opt` will be persistent. You can use the `ros console switch` command to switch to a [persistent console]({{site.baseurl}}/os/configuration/custom-console/#console-persistence) and replace the native Busybox console. For example, to switch to the Ubuntu console:

```
$ sudo ros console switch ubuntu
```

### Conclusion

RancherOS is a simple Linux distribution ideal for running Docker.  By embracing containerization of system services and leveraging Docker for management, RancherOS hopes to provide a very reliable, and easy to manage OS for running containers. 

