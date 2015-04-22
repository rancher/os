---
title: Adding System Services
layout: default

---

## Adding System Services
---

_Available as of v0.3.0._

You can design any system services in the [docker compose](https://docs.docker.com/compose/) format. The service options that are automatically available are saved in the [os-services repo](https://github.com/rancherio/os-services).

Our first example of a service is the Ubuntu console. Here's the yaml file saved in our repo.

```yaml
console:
  image: rancher/ubuntuconsole:v0.3.0
  privileged: true
  links:
  - cloud-init
  labels:
  - io.rancher.os.scope=system
  volumes_from:
  - all-volumes
  restart: always
  pid: host
  ipc: host
  net: host
```

### System-Docker vs. User-Docker

RancherOS uses labels to determine if the container should be deployed in system-docker. By default without the label, the container will be deployed in user-docker.

```yaml
labels:
  - io.rancher.os.scope=system
```
### Build

RancherOS does not support the `build` command as we are not attempting to build or rebuild services.

### Links

We use [links](https://docs.docker.com/compose/yml/#links) to link containers in another service. In our `ubuntu-console.yml`, we link the container with cloud-init, so that the console is able to use cloud-init.

```yaml
links:
  - cloud-init
```

Other examples of `links`, which use `network`, to get access to the networking container.

```yaml
links:
  - network
```

### Environment

With [environment](https://docs.docker.com/compose/yml/#environment) in the yaml file, if the environment is not set (i.e. it doesn't have an `=`), then RancherOS looks up the value in the `rancher.yml` file. 

We support worldwide globbing, so in our example below, the services.yml file will find ETCD_DISCOVERY in the `rancher.yml` file and set the environment to `https://discovery.etcd.io/d1cd18f5ee1c1e2223aed6a1734719f7` for the service. 

`services.yml` File:

```yaml
etcd:
  environment:
  - ETCD_*
```

`rancher.yml` File:

```yaml
rancher:
  environment:
  ETCD_DISCOVERY: https://discovery.etcd.io/d1cd18f5ee1c1e2223aed6a1734719f7
```
## Using System Services
---
We control system services using [rancherctl service]({{site.baseurl}}/docs/rancherctl/service/). 

To use a system service, just run `rancherctl service enable <system-service-name>` to turn on the service. By using this command, the service will also be added to the `rancher.yml` file and set to enabled, but a reboot needs to occur in order for it take effect. In the future, the reboot will be dynamic. 

The `<system-service-name>` can either be a http(s) url, location to a yaml file, or  a service that is already in the [os-services repository](https://github.com/rancherio/os-services).

Here's how we enable the ubuntu-console, which is in the os-services directory:

```bash
$ sudo rancherctl service list
disabled ubuntu-console
$ sudo rancherctl service enable ubuntu-console
$ sudo rancherctl service list
enabled ubuntu-console
$ sudo reboot
```

After the reboot and logging back in, you should be running the ubuntu console instead of the default busybox console. 

If you are using the location to a yaml file, the file must be located in `/var/lib/rancher/conf/` and the `<system-service-name>` must be `/var/lib/rancher/conf/example.yml`

Here's how we enable a service file called `example.yml`. The service file must be saved in `/var/lib/rancher/conf/`.


```bash
$ sudo rancherctl service enable /var/lib/rancher/conf/example.yml
$ sudo rancherctl service list
enabled ubuntu-console
enabled /var/lib/rancher/conf/example.yml
$ sudo reboot
```

To turn off a system service, run `rancherctl service disable <system-service-name>`. This will only turn off the service in the `rancher.yml` file, but it will not remove the service from it. Similar to when we enabled the service, we'll need to reboot in order for the disabling to take effect.

Note: Currently, there is no way to remove a service, but it's something that we will be implementing.

## Contributing to OS-Services
---
If you're interested in adding more services to RancherOS, please contribute to our [repo](https://github.com/rancherio/os-services). 

<br>
<br>