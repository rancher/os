---
title: Configuring RancherOS Networking
layout: default

---

## Configuring RancherOS Networking
---

RancherOS provides very basic support to get networking up. Changes to the networking is quite simple. 

You can change the networking settings by using `ros config` to set different keys within the network key. Anything set using this command will have its change saved in the `rancher.yml` file. Changes will only take affect after you reboot.

Alternatively, you can use a [cloud config]({{site.baseurl}}/docs/cloud-config) file to set up how the network is configured. 

To learn more information about configuring the networking settings by using `ros config`, please refer to our [cloud config]({{site.baseurl}}/docs/cloud-config) docs. 

### DNS

In the DNS section, you can set the `nameserver`, `search`, and `domain`, which directly map to the fields of the same name in `/etc/resolv.conf`.

```bash
$ sudo ros config set network.dns.domain myexampledomain.com
$ sudo ros config get network.dns.domain
myexampledomain.com
```


### Interfaces

In the `interfaces` section, the keys are used to match the desired interface to configure.  Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  The available options you can set are `address`, `gateway`, `mtu`, and `dhcp`.

```bash
$ sudo ros config set network.interfaces.eth1.address 172.68.1.0/100
$ sudo ros config get network.interfaces.eth1.address
172.68.1.0/100
```

### Multiple NICs

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

