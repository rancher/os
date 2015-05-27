---
title: Rancher.yml
layout: default

---

## Rancher.yml
---

Within RancherOS, there are various areas within the system that you might want to configure. By making changes using the `ros config` commands, you'll editing the `rancher.yml` file. Below, we'll outline the different keys that can be changed. 

`ros` is the main command to interact with RancherOS configuration, here's the link to the [full ros config command docs]({{site.baseurl}}/docs/rancheros-tools/ros/config/). With these commands, you can get and set values in the `rancher.yml` file as well as import/export configurations.

_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

## Keys in RancherOS
---


### Networking
---
RancherOS provides very basic support to get networking up. To find the various ways to change the networking, please refer to the [networking section]({{site.baseurl}}/docs/configuration/networking).


Here's the default `networking` key and the other keys within networking that can be changed.

```yaml
network:
  dns:
    nameservers: 
      - 8.8.8.8
      - 8.8.4.4
    search:
      - mydomain.com
      - example.com
    domain: mydomain.com    
  interfaces:
    eth*: {}
    eth0:
      dhcp: true
    eth1:
      match: eth1
      address: 172.19.8.101/24
      gateway: 172.19.8.1
      mtu: 1460
    lo:
      address: 127.0.0.1/8
```

#### DNS

In the DNS section, you can set the `nameserver`, `search`, and `domain`, which directly map to the fields of the same name in `/etc/resolv.conf`.

#### Interfaces

In the `interfaces` section, the keys are used to match the desired interface to configure.  Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  The available options you can set are `address`, `gateway`, `mtu`, and `dhcp`.

### Cloud Init
---

You can configure which datasources to use for cloud-init.  Multiple datasources can be set, but the datasource that is available the fastest will be used.  This value is usually pre-populated with the current setting for your environment.  Valid value are:

1. `configdrive:PATH` - Look for an OpenStack compatible config drive mounted at `PATH`
1. `file:PATH` - Read the `FILE` as the user data.
1. `ec2` - Look for EC2 style meta data at 169.254.169.254
1. `ec2:IP_ADDRESS` - Look for EC2 style meta data at the `IP_ADDRESS`
1. `url:URL` - Download `URL` and use that as the user data
1. `cmdline:URL` - Look for `cloud-config-url=URL` in `/proc/cmdline` and download `URL` as user data

Within the `cloud-init` key, you can define the datasources.

```yaml
cloud_init:
  datasources:
    - configdrive:/media/config-2
```

### Persistence
---

RancherOS will store its state in a single partition specified by the `dev` field.  The field can be a device such as `/dev/sda1` or a logical name such `LABEL=state` or `UUID=123124`.  The default value is `LABEL=RANCHER_STATE`.  The file system type of that partition can be set to `auto` or a specific file system type such as `ext4`.

```yaml
state:
 fstype: auto
 dev: LABEL=RANCHER_STATE
 autoformat:
   - /dev/sda
   - /dev/vda
```

#### Auto formatting

You can specify a list of devices to check to format on boot.  If the state partition is already found, RancherOS will not try to auto format a partition.  If the device specified in `autoformat` starts with 1 megabyte of zeros, RancherOS will autoformat the partition to ext4.  Auto-formatting is off by default.

### Upgrades
---

In the `upgrade` key, the `url` is used to find the list of available and current versions of RancherOS.

```yaml
upgrade:
  url: https://releases.rancher.com/rancheros/versions.yml
  image: rancher/os
```


### User Docker Configuration
---

The `user_docker` key configures the user-docker arguments and TLS settings.

```yaml
user_docker:
  tls: false
  tls_args: [--tlsverify, --tlscacert=ca.pem, --tlscert=server-cert.pem, --tlskey=server-key.pem,
    '-H=0.0.0.0:2376']
  args: [docker, -d, -s, overlay, -G, docker, -H, 'unix:///var/run/docker.sock']

```


### System Docker Configuration
---

The `system docker` key configures the system-docker arguments.

```yaml
system_docker:
  args: [docker, -d, --log-driver, syslog, -s, overlay, -b, docker-sys, --fixed-cidr,
    172.18.42.1/16, --restart=false, -g, /var/lib/system-docker, -G, root, -H, 'unix:///var/run/system-docker.sock']
```

## Full Export of Configuration
---

The full configurartion, including the built in default values in RancherOS can be viewed by running `ros config export --full`.  Below is the output of the full yaml file.

Any keys in the default configuration can be changed using the `ros config` command. All changes will be saved in the `rancher.yml` file.

```yaml
bootstrap_containers:
  udev:
    image: udev
    labels:
      - io.rancher.os.detach=false
      - io.rancher.os.scope=system
    log_driver: json-file
    net: host
    privileged: true
    volumes:
      - /dev:/host/dev
      - /lib/modules:/lib/modules
      - /lib/firmware:/lib/firmware
bootstrap_docker:
  args: [docker, -d, -s, overlay, -b, none, --restart=false, -g, /var/lib/system-docker,
    -G, root, -H, 'unix:///var/run/system-docker.sock']
cloud_init:
  datasources:
    - file:/var/lib/rancher/conf/user_config.yml
services_include:
  ubuntu-console: false
network:
  dns:
    nameservers: [8.8.8.8, 8.8.4.4]
  interfaces:
    eth*: {}
    eth0:
      dhcp: true
    eth1:
      match: eth1
      address: 172.19.8.101/24
    lo:
      address: 127.0.0.1/8
repositories:
  core:
    url: https://raw.githubusercontent.com/rancherio/os-services/master
state:
  fstype: auto
  dev: LABEL=RANCHER_STATE
system_containers:
  all-volumes:
    image: state
    labels:
      - io.rancher.os.createonly=true
      - io.rancher.os.scope=system
    log_driver: json-file
    net: none
    privileged: true
    read_only: true
    volumes_from:
      - docker-volumes
      - command-volumes
      - user-volumes
      - system-volumes
  cloud-init:
    image: cloudinit
    labels:
      - io.rancher.os.reloadconfig=true
      - io.rancher.os.detach=false
      - io.rancher.os.scope=system
    links:
      - cloud-init-pre
      - network
    net: host
    privileged: true
    volumes_from:
      - command-volumes
      - system-volumes
  cloud-init-pre:
    environment:
      - CLOUD_INIT_NETWORK=false
    image: cloudinit
    labels:
      - io.rancher.os.reloadconfig=true
      - io.rancher.os.detach=false
      - io.rancher.os.scope=system
    net: host
    privileged: true
    volumes_from:
      - command-volumes
      - system-volumes
  command-volumes:
    image: state
    labels:
      - io.rancher.os.createonly=true
      - io.rancher.os.scope=system
    log_driver: json-file
    net: none
    privileged: true
    read_only: true
    volumes:
      - /init:/sbin/halt:ro
      - /init:/sbin/poweroff:ro
      - /init:/sbin/reboot:ro
      - /init:/sbin/shutdown:ro
      - /init:/sbin/netconf:ro
      - /init:/usr/bin/cloud-init:ro
      - /init:/usr/bin/ros:ro
      - /init:/usr/bin/respawn:ro
      - /init:/usr/bin/system-docker:ro
      - /init:/usr/sbin/wait-for-docker:ro
      - /lib/modules:/lib/modules
      - /usr/bin/docker:/usr/bin/docker:ro
  console:
    image: console
    labels:
      - io.rancher.os.scope=system
      - io.rancher.os.remove=true
    links:
      - cloud-init
    net: host
    pid: host
    ipc: host
    privileged: true
    restart: always
    volumes_from:
      - all-volumes
  docker-volumes:
    image: state
    labels:
      - io.rancher.os.createonly=true
      - io.rancher.os.scope=system
    log_driver: json-file
    net: none
    privileged: true
    read_only: true
    volumes:
      - /var/lib/rancher:/var/lib/rancher
      - /var/lib/docker:/var/lib/docker
      - /var/lib/system-docker:/var/lib/system-docker
  network:
    image: network
    labels:
      - io.rancher.os.detach=false
      - io.rancher.os.scope=system
    links:
      - cloud-init-pre
    net: host
    privileged: true
    volumes_from:
      - command-volumes
      - system-volumes
  ntp:
    image: ntp
    labels:
      - io.rancher.os.scope=system
    links:
      - cloud-init
      - network
    net: host
    privileged: true
  syslog:
    image: syslog
    labels:
      - io.rancher.os.scope=system
    log_driver: json-file
    net: host
    privileged: true
    volumes_from:
      - system-volumes
  system-volumes:
    image: state
    labels:
      - io.rancher.os.createonly=true
      - io.rancher.os.scope=system
    log_driver: json-file
    net: none
    privileged: true
    read_only: true
    volumes:
      - /dev:/host/dev
      - /var/lib/rancher/conf:/var/lib/rancher/conf
      - /etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt.rancher
      - /lib/modules:/lib/modules
      - /lib/firmware:/lib/firmware
      - /var/run:/var/run
      - /var/log:/var/log
  udev:
    environment:
      - DAEMON=true
    image: udev
    labels:
      - io.rancher.os.detach=true
      - io.rancher.os.scope=system
    net: host
    privileged: true
    volumes_from:
      - system-volumes
  user-volumes:
    image: state
    labels:
      - io.rancher.os.createonly=true
      - io.rancher.os.scope=system
    log_driver: json-file
    net: none
    privileged: true
    read_only: true
    volumes:
      - /home:/home
      - /opt:/opt
  userdocker:
    image: userdocker
    labels:
      - io.rancher.os.scope=system
    links:
      - network
    net: host
    pid: host
    ipc: host
    privileged: true
    volumes_from:
      - all-volumes
  userdockerwait:
    image: userdockerwait
    labels:
      - io.rancher.os.detach=false
      - io.rancher.os.scope=system
    links:
      - userdocker
    net: host
    volumes_from:
      - all-volumes
system_docker:
  args: [docker, -d, --log-driver, syslog, -s, overlay, -b, docker-sys, --fixed-cidr,
    172.18.42.1/16, --restart=false, -g, /var/lib/system-docker, -G, root, -H, 'unix:///var/run/system-docker.sock']
upgrade:
  url: https://releases.rancher.com/os/versions.yml
  image: rancher/os
user_docker:
  tls_args: [--tlsverify, --tlscacert=ca.pem, --tlscert=server-cert.pem, --tlskey=server-key.pem,
    '-H=0.0.0.0:2376']
  args: [docker, -d, -s, overlay, -G, docker, -H, 'unix:///var/run/docker.sock']

```
