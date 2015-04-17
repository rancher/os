---
title: RancherCTL Service
layout: default

---

## RancherCTL Service
---
_Available as of v0.3.0._


`rancherctl service` allows you to enable/disable different system services. Please go to our [Adding System Services page]({{site.baseurl}}/docs/system-services/) for more details on how to add system services to RancherOS. The `rancherctl service` command shows you how to turn on and off the services that have been added.

### Sub Commands
---
|Command | Description |
|--------|-------------|
|`enable`	| Turn on an service|
|`disable`	|Turn off an service|
|`list`	|	List services and state|

### List
---
The `list` command will provide you a list of all services available in the [os-services repository](https://github.com/rancherio/os-services) as well as any service that was added by the user to `rancher.yml`. The command will also show the state that each service is in.

```bash
$ sudo rancherctl service list
disabled ubuntu-console
```

### Enable
---
The `enable` command turns on a service. This service can either be a http(s) url, location to a yaml file, or  a service that is already in the [os-services repository](https://github.com/rancherio/os-services). For anything outside of the os-services repo, an additional item will be added to the `rancher.yml` file. In order for the change to take effect, you must reboot. In the future, the reboot will be dynamic.

For our example, we're enabling the ubuntu console. After the reboot, we'll be logged in using the ubuntu-console. 

```bash
$ sudo rancherctl service list
disabled ubuntu-console
$ sudo rancherctl service enable ubuntu-console
$ sudo rancherctl service list
enabled ubuntu-console
$ sudo reboot
```

### Disable
---
The `disable` command turns off any service, but the service will **not** be removed. You will need to reboot in order for the change to take effect. In the future, the reboot will be dynamic. 

For our example, we're disabling the ubuntu console. After the reboot, we'll be logged in using the busybox console.

```bash
$ sudo rancherctl service list
enabled ubuntu-console
$ sudo rancherctl service disable ubuntu-console
$ sudo rancherctl service list
disabled ubuntu-console
```

