---
title: Configuring Network Interfaces in RancherOS
layout: os-default
redirect_from:
  - os/configuration/networking/
---

## Interfaces

Using `ros config`, you can configure specific interfaces. Wildcard globbing is supported so `eth*` will match `eth1` and `eth2`.  The available options you can configure are `address`, `gateway`, `mtu`, and `dhcp`.

```
$ sudo ros config set rancher.network.interfaces.eth1.address 172.68.1.100/24
$ sudo ros config set rancher.network.interfaces.eth1.gateway 172.68.1.1
$ sudo ros config set rancher.network.interfaces.eth1.mtu 1500
$ sudo ros config set rancher.network.interfaces.eth1.dhcp false
```

If you wanted to configure the interfaces through the cloud config file, you'll need to place interface configurations within the `rancher` key.

```yaml
#cloud-config
rancher:
  network:
    interfaces:
      eth1:
        address: 172.68.1.100/24
        gateway: 172.68.1.1
        mtu: 1500
        dhcp: false
```

### Multiple NICs

If you want to configure one of multiple network interfaces, you can specify the MAC address of the interface you want to configure.

Using `ros config`, you can specify the MAC address of the NIC you want to configure as follows:

```
$ sudo ros config set rancher.network.interfaces.”mac=ea:34:71:66:90:12:01”.dhcp true
```

Alternatively, you can place the MAC address selection in your cloud config file as follows:

```yaml
#cloud-config
rancher:
  network:
    interfaces:
      "mac=ea:34:71:66:90:12:01":
         dhcp: true
```

### NIC bonding

You can aggregate several network links into one virtual link for redundancy and increased throughput. For example:

```yaml
#cloud-config
rancher:
  network:
    interfaces:
      bond0:
        addresses:
        - 192.168.101.33/31
        - 10.88.23.129/31
        gateway: 192.168.101.32
        bond_opts:
          downdelay: "200"
          lacp_rate: "1"
          miimon: "100"
          mode: "4"
          updelay: "200"
          xmit_hash_policy: layer3+4
        post_up:
        - ip route add 10.0.0.0/8 via 10.88.23.128
      mac=0c:c4:d7:b2:14:d2:
        bond: bond0
      mac=0c:c4:d7:b2:14:d3:
        bond: bond0
```

In this example two physical NICs (with MACs `0c:c4:d7:b2:14:d2` and `0c:c4:d7:b2:14:d3`) are aggregated into a virtual one `bond0`.

### VLANS

In this example, you can create an interface `eth0.100` which is tied to VLAN 100 and an interface `foobar` that will be tied to VLAN 200.

```
#cloud-config
rancher:
  network:
    interfaces:
      eth0:
        vlans: 100,200:foobar
```
