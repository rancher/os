---
title: Running RancherOS on Digital Ocean
layout: os-default

---

## Running RancherOS on DigitalOcean
---

Running RancherOS on DigitalOcean is not yet supported, but there is a `rancheros` image now available from the commandline tool, so you can run:

```
$ doctl.exe compute droplet create --image rancheros --region sfo1 --size 2gb --ssh-keys 0a:db:77:92:03:b5:b2:94:96:d0:92:6a:e1:da:cd:28 myrancherosvm
ID          Name       Public IPv4    Private IPv4    Public IPv6    Memory    VCPUs    Disk    Region    Image                                    Status    Tags
47145723    myrancherosvm                                            2048      2        40      sfo1      RacherOS v1.0.1-rc [UNSUPPORTED/BETA]    new

$ doctl.exe compute droplet list
47145723    myrancherosvm                    107.170.203.111    10.134.26.83     2604:A880:0001:0020:0000:0000:2750:0001    2048      2        40      sfo1      RacherOS v1.0.1-rc [UNSUPPORTED/BETA]    active

ssh -i ~/.ssh/Sven.pem rancher@107.170.203.111
```

