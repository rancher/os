---
title: RancherCTL Services
layout: default

---

## RancherCTL Services

As of version **v0.3.0**, we've added this functionality.


`rancherctl service` allows you to add/remove different system services. Please go to our [Adding System Services page]({{site.baseurl}}/docs/system-services/) for more details.

### Sub Commands

|Command | Description |
|--------|-------------|
|`enable`	| Turn on an service|
|`disable`	|Turn off an service|
|`list`	|	List services and state|

### Enable

The `enable` command adds a service and turns it on.


### Disable

The `disable` command turns off any service that you have enabled. After being disabled, the service will **not** be removed. 


### List

The `list` command will provide you a list of all services and the state that they are in.