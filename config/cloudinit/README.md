**NOTE**: This project has been superseded by [Ignition][ignition] and is no longer under active development. Please direct all development efforts to Ignition.

[ignition]: https://github.com/coreos/ignition

# coreos-cloudinit [![Build Status](https://travis-ci.org/coreos/coreos-cloudinit.png?branch=master)](https://travis-ci.org/coreos/coreos-cloudinit)

coreos-cloudinit enables a user to customize CoreOS machines by providing either a cloud-config document or an executable script through user-data.

## Configuration with cloud-config

A subset of the [official cloud-config spec][official-cloud-config] is implemented by coreos-cloudinit.
Additionally, several [CoreOS-specific options][custom-cloud-config] have been implemented to support interacting with unit files, bootstrapping etcd clusters, and more.
All supported cloud-config parameters are [documented here][all-cloud-config]. 

[official-cloud-config]: http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data
[custom-cloud-config]: https://github.com/rancher/os/config/cloudinit/blob/master/Documentation/cloud-config.md#coreos-parameters
[all-cloud-config]: https://github.com/rancher/os/config/cloudinit/tree/master/Documentation/cloud-config.md

The following is an example cloud-config document:

```
#cloud-config

coreos:
    units:
      - name: etcd.service
        command: start

users:
  - name: core
    passwd: $1$allJZawX$00S5T756I5PGdQga5qhqv1

write_files:
  - path: /etc/resolv.conf
    content: |
        nameserver 192.0.2.2
        nameserver 192.0.2.3
```

## Executing a Script

coreos-cloudinit supports executing user-data as a script instead of parsing it as a cloud-config document.
Make sure the first line of your user-data is a shebang and coreos-cloudinit will attempt to execute it:

```
#!/bin/bash

echo 'Hello, world!'
```

## user-data Field Substitution

coreos-cloudinit will replace the following set of tokens in your user-data with system-generated values.

| Token         | Description |
| ------------- | ----------- |
| $public_ipv4  | Public IPv4 address of machine |
| $private_ipv4 | Private IPv4 address of machine |

These values are determined by CoreOS based on the given provider on which your machine is running.
Read more about provider-specific functionality in the [CoreOS OEM documentation][oem-doc].

[oem-doc]: https://coreos.com/docs/sdk-distributors/distributors/notes-for-distributors/

For example, submitting the following user-data...

```
#cloud-config
coreos:
    etcd:
        addr: $public_ipv4:4001
        peer-addr: $private_ipv4:7001
```

...will result in this cloud-config document being executed:

```
#cloud-config
coreos:
    etcd:
        addr: 203.0.113.29:4001
        peer-addr: 192.0.2.13:7001
```

## Bugs

Please use the [CoreOS issue tracker][bugs] to report all bugs, issues, and feature requests.

[bugs]: https://github.com/coreos/bugs/issues/new?labels=component/cloud-init

