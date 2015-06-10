---
title: ROS ENV
layout: default

---

## ROS ENV
---
_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

`ros env` runs an arbitrary command with RancherOS environment. 

Remember all `ros` commands needs to be used with `sudo` (or as `root` user). 

### Example
---
Suppose there are a few environment entries in `rancher.environment`: 

```sh
$ sudo ros config get environment 
ETCD_DISCOVERY: https://discovery.etcd.io/a65ee41d4a85795676f8c2d564697dcf
FLANNEL_NETWORK: 10.244.0.0/16
NODE_IP: 10.0.2.15
```

Let's run a command with that environment using `ros env`: 

```sh
$ sudo ros env sh -c 'echo $ETCD_DISCOVERY'
https://discovery.etcd.io/a65ee41d4a85795676f8c2d564697dcf
```

Normally, running `sh -c 'echo $ETCD_DISCOVERY'` wouldn't print anything 
(unless, of course, you had `ETCD_DISCOVERY` in your shell environment by any chance).

Let's try something else:

```sh
$ sudo ros config set environment.ETCD_NAME RancherOS_etcd_node
$ sudo ros env sh -c 'echo $ETCD_NAME'
RancherOS_etcd_node
```

Let's run _etcd_ with a modified environment:

```sh
$ sudo ros env system-docker run --rm -e ETCD_NAME quay.io/coreos/etcd:v2.0.10
2015/04/30 12:03:01 etcd: no data-dir provided, using default data-dir ./RancherOS_etcd_node.etcd
<skip>
2015/04/30 12:03:01 etcdserver: datadir is valid for the 2.0.1 format
2015/04/30 12:03:01 etcdserver: name = RancherOS_etcd_node
2015/04/30 12:03:01 etcdserver: data dir = RancherOS_etcd_node.etcd
<skip>
```

<br>



