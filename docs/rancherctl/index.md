---
title: RancherCTL
layout: default
---

## RancherCTL

A useful command that can be used with RancherOS is `rancherctl` which can be used to control and configure the system. `rancherctl` requires you to be the root user, so with the rancher user, you will need to use `sudo`.


### Sub Commands
| Command  | Description                                     |
|----------|-------------------------------------------------|
|`config`, `c`  |	[Configure Settings]({{site.baseurl}}/docs/rancherctl/config/)                       |
|`service`, `s`   |	[Service Settings]({{site.baseurl}}/docs/rancherctl/services/)                          |
|`os`           |   [Operating System Upgrade/Downgrade]({{site.baseurl}}/docs/rancherctl/os/)      |
|`tls`          |	[Setup TLS configuration]({{site.baseurl}}/docs/rancherctl/tls/)                 |
|`help`, `h`    |	Shows a list of commands or help for one command |


### RancherOS Version

If you want to check what version you are on, just use the `-v` option. 

```sh
$ sudo rancherctl -v
rancherctl version v0.2.1
```
### Help

To list available commands, run any `rancherctl` command with `-h` or `--help`. This would work with any subcommand within `rancherctl`.

```sh
$ sudo rancherctl -h
NAME:
rancherctl - Control and configure RancherOS

USAGE:
rancherctl [global options] command [command options] [arguments...]

VERSION:
v0.2.1

AUTHOR(S): 
Rancher Labs, Inc.  

COMMANDS:
config, c	configure settings
addon, a	addon settings
os		operating system upgrade/downgrade
tls		setup tls configuration
help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
--help, -h			show help
--generate-bash-completion	
--version, -v		print the version
```

