---
title: Configuring RancherOS
layout: default

---

# Configuring RancherOS

The configuration of RancherOS is on a single configuration file called rancher.yml.  You can either use [rancherctl config]({{site.baseurl}}/docs/rancherctl/config/) to edit and interact with this file or edit `/var/lib/rancher/conf/rancher.yml` directly.  

Note: We recommend using [rancherctl config]({{site.baseurl}}/docs/rancherctl/config/) as it is safer to use.

## rancherctl

[`rancherctl`]({{site/baseurl}}/docs/rancherctl/) is the main command to interact with RancherOS configuration, to the the [full documentation]({{site.baseurl}}/docs/rancherctl/config/).

## Networking

RancherOS provides very basic support to get networking up.

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
eth*:
dhcp: true
eth1:
address: 192.168.0.5
gateway: 192.168.0.1
mtu: 1460
lo:
address: 127.0.0.1/8
```

### DNS

In the DNS section you can set the `nameserver`, `search`, and `domain` which directly map to the fields of the same name in `/etc/resolv.conf`

### Interfaces

In the `interfaces` section the keys are used to match the desire interface to configure.  Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  The available options you can set are `address`, `gateway`, `mtu`, and `dhcp`.


## Persistence

```yaml
state:
fstype: auto
dev: LABEL=RANCHER_STATE
autoformat:
- /dev/sda
- /dev/vda
```

RancherOS will store its state in a single partition specified by the `dev` field.  The field can be a device such as `/dev/sda1` or a logical name such `LABEL=state` or `UUID=123124`.  The default value is `LABEL=RANCHER_STATE`.  The file system type of that partition can be set to `auto` or a specific file system type such as `ext4`.

### Auto formatting

You can specify a list of devices to check to format on boot.  If the state partition is already found RancherOS will not try to auto format a partition.  If the device specified in `autoformat` starts with 1 megabyte of zeros, RancherOS will autoformat the partition to ext4.  Auto-formatting is off by default.

## Upgrades

```yaml
upgrade:
url: https://releases.rancher.com/rancheros/versions.yml
```

The `url` is used to find the list of available and current versions of RancherOS.

## User Docker Configuration

```yaml
user_docker:
tls: false
tls_args: [--tlsverify, --tlscacert=ca.pem, --tlscert=server-cert.pem, --tlskey=server-key.pem,
'-H=0.0.0.0:2376']
args: [docker, -d, -s, overlay, -G, docker, -H, 'unix:///var/run/docker.sock']
```

Configure the user Docker arguments and TLS settings.

## System Docker Configuration

```yaml
system_docker:
args: [docker, -d, -s, overlay, -b, none, --restart=false, -g, /var/lib/system-docker,
-H, 'unix:///var/run/system-docker.sock']
```

## Full rancher.yml

The full rancher.yml, including the built in default values in RancherOS can be viewed by running `rancherctl config export --full`.  Below is the output of the full yaml file.


```yaml
addons:
ubuntu-console:
system_containers:
- id: console
run: --name=ubuntu-console -d --rm --privileged --volumes-from=all-volumes --restart=always
--ipc=host --net=host --pid=host rancher/ubuntuconsole:v0.0.2
bootstrap_containers:
- id: udev
run: --name=udev --net=none --privileged --rm -v=/dev:/host/dev -v=/lib/modules:/lib/modules:ro
udev
cloud_init:
datasources:
- configdrive:/media/config-2
network:
dns:
nameservers: [8.8.8.8, 8.8.4.4]
interfaces:
eth*:
dhcp: true
lo:
address: 127.0.0.1/8
state:
fstype: auto
dev: LABEL=RANCHER_STATE
autoformat:
- /dev/sda
- /dev/vda
system_docker:
args: [docker, -d, -s, overlay, -b, none, --restart=false, -g, /var/lib/system-docker,
-H, 'unix:///var/run/system-docker.sock']
upgrade:
url: https://cdn.rancher.io/rancheros/versions.yml
user_docker:
tls_args: [--tlsverify, --tlscacert=ca.pem, --tlscert=server-cert.pem, --tlskey=server-key.pem,
'-H=0.0.0.0:2376']
args: [docker, -d, -s, overlay, -G, docker, -H, 'unix:///var/run/docker.sock']
```
