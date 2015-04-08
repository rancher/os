---
title: Getting Started Guide
layout: default

---

# RancherOS Getting Started Guide

RancherOS is a lightweight Linux distro designed from the ground up to run Docker containers. This guide provides detailed instructions on how to get started using RancherOS.

## Architecture
Details on the RancherOS Architecture can be found [here]({{site.baseurl}}/docs/architecture/).

## Installing RancherOS

There are a couple of options of how to install RancherOS.

* Installing on [Vagrant]({{site.baseurl}}/docs/getting-started/vagrant/).
* Running on [Amazon]({{site.baseurl}}/docs/getting-started/amazon/).
* Installing to [Disk]({{site.baseurl}}/docs/getting-started/disk/).

## A First Look At RancherOS

For the rest of the guide, I will use Vagrant locally to run RancherOS. Let’s run RancherOS and get a first look on the system and discover how to use it.

First to login to the system, use Vagrant ssh to authenticate with the private Vagrant key and login to the system:

```sh
$ vagrant ssh
[rancher@rancher ~]$
```

To see what processes are running on the system:

```sh
[rancher@rancher ~]$ ps aux
PID   USER 	COMMAND
1 	root 	docker -d -s overlay -b none --restart=false -H unix:///var/run/system-docker.sock
…..
308     root 	ntpd -d
314     root 	rsyslogd -n
322     root 	docker -d -s overlay --tlsverify --tlscacert=/etc/docker/tls/ca.pem --tlscert=/etc/docker/tls/serve
…..

```

As you can see that the first process on the system, PID1, is the Docker daemon we talked about earlier, called **system-docker**. This is where RancherOS runs system services like ntpd and rsyslogd. You can use the system-docker command to control the **system-docker** daemon. 

The other Docker daemon running on the system is **user-docker**, which can be accessed by using the normal docker command.

Use docker images command to see the images that the system has:

```sh
[rancher@rancher ~]$ docker images
REPOSITORY   TAG	IMAGE ID	CREATED	VIRTUAL SIZE
```

At this point there are no containers runing on the user-docker daemon. However, if  you run the same command against the sytem-docker instance you’ll see a number of system services that are shipped with RancherOS. [Note: system-docker can only be used by root, so it is necessary to use the sudo command whenever you want to interact with system-docker]:

```
[rancher@rancher ~]$ sudo system-docker images
REPOSITORY  TAG         IMAGE ID        CREATED     	VIRTUAL SIZE
syslog  	latest  	92855074bb56    46 hours ago    18.09 MB
syslog      v0.0.1      92855074bb56    46 hours ago    18.09 MB
ntp         latest      6560c12b3f56    46 hours ago    18.09 MB
ntp         v0.0.1      6560c12b3f56    46 hours ago    18.09 MB
rescue      latest      27c5b8ae9b7c    46 hours ago    18.1 MB
rescue      v0.0.1      27c5b8ae9b7c    46 hours ago    18.1 MB
console     latest      ee3a47bd7309    46 hours ago    18.1 MB
console     v0.0.1      ee3a47bd7309    46 hours ago	18.1 MB
userdocker  latest  	c0196a52a7c4    46 hours ago    18.35 MB
userdocker  v0.0.1      c0196a52a7c4    46 hours ago    18.35 MB
cloudinit   latest      7dc2bc8c2ad5	46 hours ago	18.09 MB
cloudinit   v0.0.1      7dc2bc8c2ad5    46 hours ago	18.09 MB
…….
```

All of these images are available for use by system docker daemon, some of them are run at boot time, and others,  such as the console, userdocker, rsyslog, and ntp containers are always running:

```
[rancher@rancher ~]$ sudo system-docker ps
CONTAINER ID    IMAGE   COMMAND      	CREATED         	STATUS         PORTS             NAMES
5ff08dbb57ce   console:latest 	"/usr/sbin/console.s   About an hour ago   Up About an hour console
56ac381d4acb   userdocker:latest   "/docker.sh"	About an hour ago   Up About an hour   userdocker    	 
f0dd31b1f7a8   syslog:latest   	"/syslog.sh"   	About an hour ago   Up About an hour    syslog         	 
0c7154630edd   ntp:latest      	"/ntp.sh"     	About an hour ago   Up About an hour    ntp
```


 Overview of how to deploy Docker containers on RancherOS.
- Instructions for deploying system containers.
- Details on how to use RancherOS with Rancher management software.

## Using RancherCTL
Another useful command that can be used with RancherOS is rancherctl which can be used to control and configure the system:

```sh
[rancher@rancher ~]$ rancherctl -v
rancherctl version 0.0.1
```

RancherOS state is controlled by simple document, rancherctl is used to edit the configuration of the system, to see for example the dns configuration of the system:

```sh
[rancher@rancher ~]$ sudo rancherctl config get dns
- 8.8.8.8
- 8.8.4.4
```

You can use rancherctl to customize the console and replace the native Busybox console with the consoles from other Linux distributions.  Initially RancherOS only supports the Ubuntu console, but other console support will be coming soon To enable the Ubuntu console use the following command:

```sh
[rancher@rancher ~]$ sudo rancherctl addon enable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

After that you will be able to use Ubuntu console, to turn it off use disable instead of enable, and then reboot.

```sh
rancher@rancher:~$ sudo rancherctl addon disable ubuntu-console;
```

Note that any changes to the console or the system containers will be lost after reboots, any changes to /home or /opt will be persistent. Theconsole always executes **/opt/rancher/bin/start.sh** at each startup. 

## Deploying a Docker Container

The next thing is to deploy a normal Docker container on the user-docker daemon.  The RancherOS User docker daemon is identical to any other Docker environment, so all normal Docker commands work.

The following is an example of deploying a small nginx container installed on a busybox linux. To start a Docker container in user docker environment, use the following command:

```sh
rancher@rancher:~$ docker run -d --name nginx -p 8000:80 husseingalal/nginxbusy
be2a3c972b75e95cd162e7b4989f66e2b0ed1cb90529c52fd93f6c849b01840f
```

You can see that the nginx container is up and running, using docker ps command:

```sh
rancher@rancher:~$ docker ps
CONTAINER ID        IMAGE                           COMMAND             CREATED             STATUS              PORTS                  NAMES
be2a3c972b75        husseingalal/nginxbusy:latest   "/usr/sbin/nginx"   3 seconds ago       Up 2 seconds        0.0.0.0:8000->80/tcp   nginx
```

Note that the rancher user belongs to docker group, thats why I‘ve been able to use docker without sudo privileges.

## Deploying A System Service Container

The following is a simple docker container to set up linux-dash a minimal low-overhead web dashboard for monitoring linux servers, the Dockerfile will be like this:

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

I used hwestphal/nodebox image which uses a busybox image and installs node.js and npm, then I downloaded the source code of linux-dash, finally I then ran the server. This Docker image is only 32MB, and linux-dash will run on port 80 by default.

To run this container with system-docker use the following command:

```sh
$ sudo system-docker run -d --net=host --name busydash husseingalal/busydash
```

Note that I used --net=host to tell system Docker not to containerize the container's networking, and use the host’s instead. After running the container you can see the monitoring server by accessing http://ip-of-the-server

![System Docker Container]({{site.baseurl}}/img/Rancher_busydash.png)

To make the container survive during the reboots, you should create the /opt/rancher/bin/start.sh script, and add the docker start line to launch the docker at each startup:

```
$ sudo mkdir -p /opt/rancher/bin
$ sudo echo “sudo system-docker start busydash” >> /opt/rancher/bin/start.sh
$ sudo chmod 755 /opt/rancher/bin/start.sh
```

## Using Rancher Management platform with RancherOS

Rancher Management platform can be used to Manage Docker containers on RancherOS machines, in the following example I am going to illustrate how to set up Rancher platform and register RancherOS installed on EC2 machine, first you need to run Rancher platform on a machine using the following command:

```
root@Rancher-mng:~# docker run -d -p 8080:8080 rancher/server
```

![Rancher Platform 1]({{site.baseurl}}/img/Rancher_platform1.png)

The next step is to register your RancherOS machine with Rancher platform by executing the following command on the command line of the RancherOS host:

```
$ docker run --rm -it --privileged -v /var/run/docker.sock:/var/run/docker.sock rancher/agent http://MANAGEMENT_IP:8080
```

You should see the RancherOS machine on the management platform:

![Rancher Platform 2]({{site.baseurl}}/img/Rancher_platform2.png)

You can now start to deploy your Docker containers on RancherOS using the Rancher management platform, pretty cool, right?

## Conclusion

RancherOS is a simple Linux distribution ideal for running Docker.  It is very new, and evolving quickly. **It is absolutely not a production quality Linux distribution at this point**. However, by embracing containerization of system services and leveraging Docker for management, RancherOS hopes to provide a very reliable, and easy to manage OS for running containers.  To stay up to date, please follow the RancherOS [GitHub site](https://github.com/rancherio/os).  
