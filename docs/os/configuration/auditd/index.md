---
title: Auditd Settings in RancherOS
layout: os-default

---

## Auditd Settings
---

The Linux auditing system is designed to be able to provide a log of events and actions on a server. This consists of two parts, the kernel audit module and the auditd daemon. You can configure whether this is enabled or not and what is audited in cloud-config. This can be overridden as required.

By default, kernel auditing is _disabled_ and has a ruleset designed to comply with the CIS Docker 1.13 Benchmark.

```
auditd:
   enabled: false
   rules:
     - "-i"
     - "-w /etc/docker -k docker"
     - "-w /usr/bin/docker-containerd-ctr -k docker"
     - "-w /usr/bin/docker-containerd -k docker"
     - "-w /usr/bin/docker-containerd-shim -k docker"
     - "-w /usr/bin/docker -k docker"
     - "-w /usr/bin/docker-runc -k docker"
     - "-w /var/lib/docker -k docker"
     - "-w /var/run/docker.sock -k docker"
 ```

 If you decide to enable kernel auditing, you should consider streaming the contents of `/var/log/audit/audit.log` to an ELK stack or SIEM for further analysis using something like fluentd. As this logfile can grow quite large, you should also run logrotate against it.

 _Nb._ New system service for this?
