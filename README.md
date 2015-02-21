# RancherOS

It's containers all the way down!  Everything is powered by Docker (I mean everything).

## Building

Docker 1.5+ required.

    ./build.sh

## Running

KVM, VirtualBox, and VMware all supported (Clouds and Vagrant coming soon).  Just
boot from the `rancheros.iso` from the [releases](https://github.com/rancherlabs/os/releases) page.

## Logging in

Log in with rancher/rancher and use `sudo` to get root access.

Once cloud-init integration is finished we will not need to hard code the
password anymore.

## Persisting State

Create a partition with the label `RANCHER_STATE`, for example

    mkfs.ext4 -L RANCHER_STATE /dev/sda

## Configuring

The entire state of RancherOS is controlled by a single configuration document.
You can edit the configuration with the `rancherctl config` command.

## Commands

`docker` -- Good old Docker, use that to run stuff.

`system-docker` -- The docker instance running the system containers.  Must run as root

`rancherctl` -- Control and configure RancherOS

## How does this work

Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as PID 1.  System Docker then launches
a container that runs the user Docker.  We created this separation as is really seemed like a
bad idea that if somebody did `docker rm -f $(docker ps -qa)` they would delete the entire OS.

![How it works](https://raw.githubusercontent.com/rancherlabs/os/master/docs/rancheros.png "How it works")

