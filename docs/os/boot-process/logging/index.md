---
title: System logging
layout: os-default

---

## RancherOS system logging

### System services

RancherOS uses containers for its system services. This means the logs for `syslog`, `acipd`, `system-cron`, `udev`, `network`, `ntp`, `console` and the user Docker are available using `sudo ros service logs <service-name>`. 

### Boot logging

Since v1.1.0, the init process's logs are copied to `/var/log/boot` after the user-space filesystem is made available. These can be used to diagnose initialisation, network, and cloud-init issues.

### Remote Syslog logging

The Linux kernel has a `netconsole` logging facility that allows it to send the Kernel level logs to a remote Syslog server.
When you set this kernel boot parameter in RancherOS v1.1.0 and later, the RancherOS debug logs will also be sent to it.

To set up Linux kernel and RancherOS remote Syslog logging, you need to set both a local, and remote host IP address - even if this address isn't the final IP address of your system. The kernel setting looks like:

```
 netconsole=[+][src-port]@[src-ip]/[<dev>],[tgt-port]@<tgt-ip>/[tgt-macaddr]

   where
        +             if present, enable extended console support
        src-port      source for UDP packets (defaults to 6665)
        src-ip        source IP to use (interface address)
        dev           network interface (eth0)
        tgt-port      port for logging agent (6666)
        tgt-ip        IP address for logging agent
        tgt-macaddr   ethernet MAC address for logging agent (broadcast)
```

For example, on my current test system, I have set the kernel boot line to:


```
printk.devkmsg=on console=tty1 rancher.autologin=tty1 console=ttyS0 rancher.autologin=ttyS0    rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda,/dev/vda] rancher.rm_usr loglevel=8 netconsole=+9999@10.0.2.14/,514@192.168.42.223/
```

The kernel boot parameters can be set during installation using `sudo ros install --append "...."`, or on an installed RancherOS system,  by running `sudo ros config syslinx` (which will start vi in a container, editing the `global.cfg` boot config file.
