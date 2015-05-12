---
title: Configuring RancherOS Networking
layout: default

---

## Configuring RancherOS Networking
---

RancherOS networking can be configured if a couple of ways.

To learn more information about configuring the networking settings by using `ros config`, please refer to our [cloud config]({{site.baseurl}}/docs/cloud-config) docs. 

### Networking
---
RancherOS provides very basic support to get networking up.

Here’s the default `networking` key and the other keys within networking that can be changed.

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

In the `interfaces` section, the keys are used to match the desired interface to configure.  Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  The available options you can set are `address`, `gateway`, `mtu`, and `dhcp`.


If you have multiple NICs on your server and you want to select a sepecific NIC for RancherOS, you will need to update the `interfaces` key. You can change this key in the [cloud config]({{site.baseurl}}/docs/cloud-config) so that it will select the NIC selection upon the first install.

```yaml
#cloud-config

#Remember, any changes for rancher will be within the rancher key
rancher:
  network:
    interfaces:
      "mac=00:00:00:00:00:00":
         dhcp: true
```

