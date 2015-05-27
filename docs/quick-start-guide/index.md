---
title: Quick Start Guide
layout: default

---

## Quick Start Guide
---

If you have a specific RanchersOS machine requirements, please check out our [guides on running RancherOS]({{site.baseurl}}/docs/running-rancheros/). With the rest of this guide, we'll start up a RancherOS using [Vagrant]({{site.baseurl}}/docs/running-rancheros/workstation/vagrant/) and show you some of what RancherOS can do.

### Launching RancherOS using Vagrant

We have created a [RancherOS Vagrant project](https://github.com/rancherio/os-vagrant) that allows you to quickly launch RancherOS on your local machine. We have a more detailed guide on [Vagrant].

1. Vagrant must be installed on your local machine. Vagrant can be downloaded and installed from [here](http://www.vagrantup.com/downloads.html).

2. Clone the [RancherOS Vagrant repository](https://github.com/rancherio/os-vagrant). Clone the repo and go into the newly cloned directory.


```bash
$ git clone https://github.com/rancherio/os-vagrant.git
$ cd os-vagrant
```

3. Startup your VM with `vagrant up`.  

```bash
$ vagrant up
Bringing machine 'rancher-01' up with 'virtualbox' provider...
…
…
==> rancher-01: Machine booted and ready!
==> rancher-01: Configuring and enabling network interfaces...
```

4. Log into your VM with `vagrant ssh`. 

```bash
$ vagrant ssh
```

With those simple commands, you're up and running a RancherOS instance.

### A First Look At RancherOS

Let's start checking out what processes are running on the system.

```sh
$ ps aux
PID   USER 	COMMAND
1 	root 	docker -d -s overlay -b none --restart=false -H unix:///var/run/system-docker.sock
…..
308     root 	ntpd -d
314     root 	rsyslogd -n
322     root 	docker -d -s overlay --tlsverify --tlscacert=/etc/docker/tls/ca.pem --tlscert=/etc/docker/tls/serve
…..

```

As you can see, the first process on the system, PID1, is the Docker daemon, called **system-docker**. This is where RancherOS runs system services like ntpd and rsyslogd. You can use the `system-docker` command to control the **system-docker** daemon. 

The other Docker daemon running on the system is **user-docker**, which can be accessed by using the normal `docker` command.

The [architecture]({{site.baseurl}}/docs/architecture/) section covers these daemons in more detail.

Use `docker images` to see the images that the system has:

```bash
$ docker images
REPOSITORY   TAG	IMAGE ID	CREATED	VIRTUAL SIZE
```

At this point, there are no containers running on the user-docker daemon. However, if you run the same command against the system-docker instance you’ll see a number of system services that are shipped with RancherOS. 

Note: system-docker can only be used by root, so it is necessary to use the sudo command whenever you want to interact with system-docker

```bash
$ sudo system-docker images
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

All of these images are available for use by system-docker daemon, some of them are run at boot time, and others, such as the console, user-docker, rsyslog, and ntp containers are always running.

```bash
$ sudo system-docker ps
CONTAINER ID    IMAGE   COMMAND      	CREATED         	STATUS         PORTS             NAMES
5ff08dbb57ce   console:latest 	"/usr/sbin/console.s   About an hour ago   Up About an hour console
56ac381d4acb   userdocker:latest   "/docker.sh"	About an hour ago   Up About an hour   userdocker    	 
f0dd31b1f7a8   syslog:latest   	"/syslog.sh"   	About an hour ago   Up About an hour    syslog         	 
0c7154630edd   ntp:latest      	"/ntp.sh"     	About an hour ago   Up About an hour    ntp
```

## Deploying a Docker Container

Let's try to deploy a normal Docker container on the user-docker daemon.  The RancherOS user-docker daemon is identical to any other Docker environment, so all normal Docker commands work.

The following is an example of deploying a small nginx container installed on a Busybox Linux. To start a Docker container in the user-docker environment, use the following command:

```bash
$ docker run -d --name nginx -p 8000:80 husseingalal/nginxbusy
be2a3c972b75e95cd162e7b4989f66e2b0ed1cb90529c52fd93f6c849b01840f
```

You can see that the nginx container is up and running, using `docker ps` command:

```sh
$ docker ps
CONTAINER ID        IMAGE                           COMMAND             CREATED             STATUS              PORTS                  NAMES
be2a3c972b75        husseingalal/nginxbusy:latest   "/usr/sbin/nginx"   3 seconds ago       Up 2 seconds        0.0.0.0:8000->80/tcp   nginx
```

Note: The rancher user belongs to docker group, which is why we're able to use docker without sudo privileges.

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

I used hwestphal/nodebox image, which uses a busybox image and installs node.js and npm. I downloaded the source code of Linux-dash, and finally I ran the server. This Docker image is only 32MB, and linux-dash will run on port 80 by default.

To run this container with system-docker use the following command:

```sh
$ sudo system-docker run -d --net=host --name busydash husseingalal/busydash
```

Note that I used --net=host to tell system-docker not to containerize the container's networking, and use the host’s networking instead. After running the container, you can see the monitoring server by accessing http://SERVER_IP.

![System Docker Container]({{site.baseurl}}/img/Rancher_busydash.png)

To make the container survive during the reboots, you should create the `/opt/rancher/bin/start.sh` script, and add the docker start line to launch the docker at each startup:

```bash
$ sudo mkdir -p /opt/rancher/bin
$ sudo echo “sudo system-docker start busydash” >> /opt/rancher/bin/start.sh
$ sudo chmod 755 /opt/rancher/bin/start.sh
```

### Using [ROS]({{site.baseurl}}/docs/rancheros-tools/ros/)
Another useful command that can be used with RancherOS is `ros` which can be used to control and configure the system. 

_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

```bash
$ ros -v
ros version 0.0.1
```

RancherOS state is controlled by simple document, which is **/var/lib/rancher/conf/rancher.yml**. `ros` is used to edit the configuration of the system, to see for example the dns configuration of the system:

```sh
$ sudo ros config get dns
- 8.8.8.8
- 8.8.4.4
```

You can use ros to customize the console and replace the native Busybox console with the consoles from other Linux distributions.  Initially, RancherOS only supports the Ubuntu console, but other console support will be coming soon. In order to enable the Ubuntu console use the following command:

```sh
$ sudo ros addon enable ubuntu-console;
$ sudo reboot
```

After that you will be able to use Ubuntu console, to turn it off use disable instead of enable, and then reboot.

```sh
$ sudo ros addon disable ubuntu-console;
```

Note that any changes to the console or the system containers will be lost after reboots, any changes to /home or /opt will be persistent. The console always executes **/opt/rancher/bin/start.sh** at each startup. 


### Using Rancher Management platform with RancherOS

Rancher Management platform can be used to Manage Docker containers on RancherOS machines, in the following example I am going to illustrate how to set up Rancher platform and register RancherOS installed on EC2 machine, first you need to run Rancher platform on a machine using the following command:

```bash
$ sudo docker run -d --restart=always -p 8080:8080 rancher/server
```

You can access the Rancher server by going to the http://SERVER_IP:8080. It might take a couple of minutes before it is available.

Note: If you are trying to use an EC2 instance, you will need to make sure the TCP port 8080 has been enabled in order to view the Rancher server UI. To do this enablement, check the security group of the EC2 instance and update the Inbound tab to add this port 8080.

![Rancher Platform 1]({{site.baseurl}}/img/Rancher_platform1.png)

The next step is to register a RancherOS machine with Rancher platform by following the UI in the Rancher server. Select the **Custom** option and get the `docker` command to run in your RancherOS. Typically, we recommend having the Rancher server and hosts be on separate VMs, but in our example, we will use the same RancherOS instance.

Note: If you are trying to use an EC2 instance, you will need to make sure the TCP ports 9345 and 9346 are enabled as well as UDP ports 500 and 4500. To do this enablement, check the security group of the EC2 instance and update the Inbound tab to add these ports.

You should see the RancherOS machine on the management platform:

![Rancher Platform 2]({{site.baseurl}}/img/Rancher_platform2.png)

You can now start to deploy your Docker containers on RancherOS using the Rancher management platform, pretty cool, right?

### Conclusion

RancherOS is a simple Linux distribution ideal for running Docker.  It is very new, and evolving quickly. **It is absolutely not a production quality Linux distribution at this point**. However, by embracing containerization of system services and leveraging Docker for management, RancherOS hopes to provide a very reliable, and easy to manage OS for running containers.  To stay up to date, please follow the RancherOS [GitHub site](https://github.com/rancherio/os).  

<br>
