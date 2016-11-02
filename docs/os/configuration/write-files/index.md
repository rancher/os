---
title: Writing Files in RancherOS
layout: os-default

---

## Writing Files
---

You can automate writing files to disk using the `write_files` cloud-config directive.

```yaml
#cloud-config
write_files:
  - path: /etc/rc.local
    permissions: "0755"
    owner: root
    content: |
      #!/bin/bash
      echo "I'm doing things on start"
```

### Writing Files in Specific System Services

By default, the `write_files` directive will create files in the console container. To write files in other system services, the `container` key can be used. For example, the `container` key could be used to write to `/etc/ntp.conf` in the NTP system service.

```yaml
#cloud-config
write_files:
  - container: ntp
    path: /etc/ntp.conf
    permissions: "0644"
    owner: root
    content: |
      server 0.pool.ntp.org iburst
      server 1.pool.ntp.org iburst
      server 2.pool.ntp.org iburst
      server 3.pool.ntp.org iburst

      # Allow only time queries, at a limited rate, sending KoD when in excess.
      # Allow all local queries (IPv4, IPv6)
      restrict default nomodify nopeer noquery limited kod
      restrict 127.0.0.1
      restrict [::1]
```

