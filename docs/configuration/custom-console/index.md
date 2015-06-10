---
title: Custom Console OS
layout: default

---

## Custom Console OS
---

By default, RancherOS starts with the ubuntu console enabled. This is a [system service]({{site.baseurl}}/docs/configuration/system-services) that has been enabled by Rancher.

RancherOS also comes with a busybox and debian console. 

You can view which console is being used by RancherOS by checking which console container is running in system-docker. 

To enable the busybox console, all consoles in the system-services list will need to be disabled. You can check which consoles are enabled from the `ros service list` command and disable any consoles that are enabled. If you disable any of the consoles, you will need to reboot in order to be using the new console.

```bash
$ sudo ros service list
disabled debian-console 
disabled ubuntu-console
```

To enable the debian console, you'll need to enable the debian console in the services list. You'll want to make sure that the ubuntu-console is also disabled. 

```bash
$ sudo ros service list
disabled debian-console 
enabled ubuntu-console
$ sudo ros service disable ubuntu-console
$ sudo ros service enable debian-console
$ sudo reboot
```

To enable the ubuntu console, you'll need to enable the ubuntu console in the services list. You'll want to make sure that the debian-console is also disabled. 

```bash
$ sudo ros service list
disabled debian-console 
enabled ubuntu-console
$ sudo ros service disable ubuntu-console
$ sudo ros service enable debian-console
$ sudo reboot
```

When multiple consoles are enabled, the first console that starts will be the console that is used in RancherOS, so it's important to disable any consoles that you don't want to use. 


