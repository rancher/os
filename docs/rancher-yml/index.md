---
title: Rancher.yml
layout: default

---

## Rancher.yml
---

The `rancher.yml` file extends and overwrites the result of cloud config, which in turn extends and overwrites RancherOS default config. You should not manually edit the `rancher.yml` file. You invoke `ros config` command to edit `rancher.yml`.

`ros` is the main command to interact with RancherOS configuration, here's the link to the [full ros config command docs]({{site.baseurl}}/docs/rancheros-tools/ros/config/). With these commands, you can get and set values in the `rancher.yml` file as well as import/export configurations.

You can view the content of `rancher.yml` file by issuing the `ros config export` command. Another command `ros config export --full` prints the current effective configuration of RancherOS, taking into account the initial default configuration and the impact of cloud config.

_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

We will now walk through various sections of `rancher.yml` file.

### Networking
---
The networking section in `rancher.yml` has identical syntax as the networking directives under the `rancher` key in cloud config files. The networking configuration in `rancher.yml`, however, will extend and overwrite the networking directives in cloud config.

RancherOS provides `ros config` commands to configure networking settings in `rancher.yml`. For details, please refer to the [networking section]({{site.baseurl}}/docs/configuration/networking).

Here is an example of network configuration section in the `rancher.yml` file:

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

In the `interfaces` section, the keys are used to match the desired interface to configure.  Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  Specific MAC address can be used to pick the NIC interface using `"mac=XXX"` as a key. The available options you can set are `address`, `gateway`, `mtu`, and `dhcp`.

### Cloud Init Data Sources
---

You can configure which data sources to use for cloud-init.  Multiple data sources can be set, but the data source that is available the fastest will be used.  This value is usually pre-populated with the current setting for your environment.  Valid value are:

1. `configdrive:PATH` - Look for an OpenStack compatible config drive mounted at `PATH`
1. `file:PATH` - Read the `FILE` as the user data.
1. `ec2` - Look for EC2 style meta data at 169.254.169.254
1. `ec2:IP_ADDRESS` - Look for EC2 style meta data at the `IP_ADDRESS`
1. `url:URL` - Download `URL` and use that as the user data
1. `cmdline:URL` - Look for `cloud-config-url=URL` in `/proc/cmdline` and download `URL` as user data

Within the `cloud-init` key, you can define the data sources.

```yaml
cloud_init:
  datasources:
    - configdrive:/media/config-2
```

### Persistent State Partition
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

The `user_docker` key configures the docker arguments and TLS settings.

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
