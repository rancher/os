---
title: Environment
layout: os-default

---

## Environment
---

The [environment key](https://docs.docker.com/compose/yml/#environment) can be used to customize system services. When a value is not assigned, RancherOS looks up the value from the `rancher.environment` key.

In the example below, `ETCD_DISCOVERY` will be set to `https://discovery.etcd.io/d1cd18f5ee1c1e2223aed6a1734719f7` for the `etcd` service.

```yaml
rancher:
  environment:
    ETCD_DISCOVERY: https://discovery.etcd.io/d1cd18f5ee1c1e2223aed6a1734719f7
  services:
    etcd:
      ...
      environment:
      - ETCD_DISCOVERY
```

Wildcard globbing is also supported. In the example below, `ETCD_DISCOVERY` will be set as in the previous example, along with any other environment variables beginning with `ETCD_`.

```yaml
rancher:
  environment:
    ETCD_DISCOVERY: https://discovery.etcd.io/d1cd18f5ee1c1e2223aed6a1734719f7
  services:
    etcd:
      ...
      environment:
      - ETCD_*
```
