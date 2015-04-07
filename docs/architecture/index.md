---
title: Ranceher OS Architecture
layout: default

---

## RancherOS Architecture

RancherOS is a 20MB distro which runs the latest Docker daemon as PID1, the first process on the system.  All other system services, like ntpd, rsyslog, and console, are running in Docker containers.

The first process on the system is called **System Docker** and it's responsible for managing the system services on RancherOS. System Docker replaces traditional init systems like systemd, and can be used to launch additional system services, which we will see later in this guide..

System Docker runs a special container called **User Docker** which is another Docker daemon responsible for managing all of the user’s containers. Any containers you launch as a user from the console will run inside this User Docker. This creates isolation from the System Docker containers, and ensures normal user commands don’t impact system services.

![GitHub Logo]({{site.baseurl}}/img/rancheroshowitworks.png)

RancherOS is dramatically smaller than most traditional operating systems, because it only includes the services necessary to run Docker.  By removing unnecessary libraries and services, requirements for security patches and other maintenance are dramatically reduced.  This is possible because with Docker, users typically package all necessary libraries into their containers.

Another way in which RancherOS is designed specifically for running Docker is that it always runs the latest version of Docker. This allows users  to take advantage of the latest Docker capabilities and bug fixes. 

Like other minimalist Linux distributions, RancherOS boots incredibly quickly, generally in 5-10 seconds.  Starting Docker containers is nearly instant, similar to starting any other process. This quickness is ideal for organizations adopting microservices and autoscaling.

Docker is an open-source platform designed for developers, system admins, and DevOps, it is used to build, ship, and run containers, using simple yet powerful CLI, you can get started with Docker from [Docker user guide](https://docs.docker.com/userguide/).