---
title: RancherCTL Service
layout: default

---

## RancherCTL Service

As of version **v0.3.0**, we've added this functionality.


`rancherctl service` allows you to add/remove different system services. Please go to our [Adding System Services page]({{site.baseurl}}/docs/system-services/) for more details.

### Sub Commands

|Command | Description |
|--------|-------------|
|`enable`	| Turn on an service|
|`disable`	|Turn off an service|
|`list`	|	List services and state|

### List

The `list` command will provide you a list of all services and the state that they are in.

```bash
$ sudo rancherctl service list
disabled ubuntu-console
```

### Enable

The `enable` command turns on a service. For our example, we're enabling the ubuntu console. For the changes to take effect, we'll need to reboot. Upon the reboot, we'll be logged in using the ubuntu-console. 

```bash
$ sudo rancherctl service list
disabled ubuntu-console
$ sudo rancherctl service enable ubuntu-console
$ sudo rancherctl service list
enabled ubuntu-console
```

### Disable

The `disable` command turns off any service, but the service will **not** be removed. For our example, we're disabling the ubuntu console. For the changes to take effect, we'll need to reboot.

```bash
$ sudo rancherctl service list
enabled ubuntu-console
$ sudo rancherctl service disable ubuntu-console
$ sudo rancherctl service list
disabled ubuntu-console
```

