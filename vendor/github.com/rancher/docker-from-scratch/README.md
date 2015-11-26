# Docker `FROM scratch`

Docker-in-Docker image based off of the empty image `scratch`.  Only the bare minimum required files are included to make Docker run.  This image weighs in around 25MB expanded.

## Running

### Overlay

```bash
# Daemon
docker run --name daemon --privileged -d rancher/docker

# Client
docker exec -it daemon docker ps
```

### Aufs

```bash
# Daemon
docker run --name daemon --privileged -d rancher/docker daemon -s aufs

# Client
docker exec -it daemon docker ps
```

## Embed in Custom Image

Since docker-from-scratch doesn't assume a base Linux distro it can be easily copied into an other image to add Docker-in-Docker capabilities

```bash
docker export $(docker create rancher/docker) > files.tar

cat > Dockerfile << EOF

FROM ubuntu
ADD files.tar /
ENTRYPOINT ["/usr/bin/dockerlaunch", "/usr/bin/docker"]
VOLUME /var/lib/docker
CMD ["daemon", "-s", "overlay"]

EOF

docker build -t custom-dind .
```

## Graph Driver Compatibility

This image is really designed to run with overlay.  Aufs is known to work but other graph drivers may not work properly or be missing userspace programs needed.


## Seriously, Why?

This code and the supporting files were extracted out of RancherOS into a separate library and are still used by RancherOS.  RancherOS runs Docker as a PID 1 but before we can exec Docker we need to setup a minimal environment for Docker in which to run.  Since RancherOS is executed by the kernel there is absolutely nothing setup in the system.  At Rancher we wrote a small amount of code to setup all the required mounts and directories to launch Docker.

We moved this code out into a separate project for two reasons.  First was simply that we wanted to clean up and modularize the RancherOS code base.  Second is that we wanted to demonstrate clearly what exactly Docker requires from the Linux user space.  For the most part Docker requires the standard mounts (`/proc`, `/sys`, `/run`, `/var/run`, etc) and the cgroup mounts in `/sys/fs/cgroup` plus the following programs/files:


```
/etc/ssl/certs/ca-certificates.crt
/usr/bin/modprobe
/usr/bin/iptables
/usr/bin/ssh
/usr/bin/xz
/usr/bin/git
/usr/bin/ps
/usr/libexec/git-core/git-clone
/usr/libexec/git-core/git-submodule
/usr/libexec/git-core/git-checkout
```

This list can be reduced to a bare minimum if you ignore certain features of Docker.  A full description of why each program is needed is below.

File | Description | Can it be ignored
-----|-------------|------------------
`/etc/ssl/certs/ca-certificates.crt` | Used as the CA roots to validate SSL connections | No
`/usr/bin/modprobe` | Used to ensure that bridge, nf_nat, br_netfilter, aufs, or overlay modules are loaded.  Additionally iptables loads kernel modules based on the configuration of the rules | Yes, just load the modules from the host that you will need.
`/usr/bin/iptables` | Docker uses IPtables to setup networking | Yes, add `--iptables=false` to the `docker -d` command.  Networking will have to be manually configured in this situation
`/usr/bin/ssh`| Used by git to clone repos over SSH | Yes, don't use git based Docker builds
`/usr/bin/xz` | Used to extract *legacy* Docker images that were compressed with xz | Yes, only use newer images.  Most popular images are not based on xz
`/usr/bin/git` | Used to do Docker builds from a git URL | Yes, don't use git based Docker builds
`/usr/bin/ps` | `docker ps` uses the host `ps` to get information about the running process in a container | No
`/usr/libexec/git-core/git-clone`| Used by git | Yes, don't use git based Docker builds
`/usr/libexec/git-core/git-submodule`| Used by git | Yes, don't use git based Docker builds
`/usr/libexec/git-core/git-checkout`| Used by git | Yes, don't use git based Docker builds

## Custom Bridge Name

If you want to run with a custom bridge name you must pass both `--bip` and `-b` as arguments.  Normally this would be an error for Docker but in this situation the docker-from-scratch container will create the bridge device with the IP address specified and then old pass `-b` to Docker.

# Troubleshooting

## Weird module loading errors

For various reasons Docker or iptables may try to load a kernel module.  You can either manually load all the needed modules from the host or you can bind mount in the kernel modules by adding `-v /lib/modules/$(uname -r)/lib/modules/$(uname -r)` to your `docker run` command

## Debug Logging

To enable debug logging on the startup of docker-from-scrach just add `-e DOCKER_LAUNCH_DEBUG=true` to the `docker run` command.  For example:

    docker run --name daemon --privileged -d -e DOCKER_LAUNCH_DEBUG=true rancher/docker
