---
title: ROS (formerly RancherCTL)
layout: default
---

## ROS
---
_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

A useful command that can be used with RancherOS is `ros` which can be used to control and configure the system. `ros` requires you to be the root user, so with the rancher user, you will need to use `sudo`.

### Sub Commands
---
| Command  | Description                                     |
|----------|-------------------------------------------------|
|`config`, `c`  |	[Configure Settings]({{site.baseurl}}/docs/rancheros-tools/ros/config/)                       |
|`env`, `e`     | [Run a command with RancherOS environment]({{site.baseurl}}/docs/rancheros-tools/ros/env/)      |
|`service`, `s`   |	[Service Settings]({{site.baseurl}}/docs/rancheros-tools/ros/service/)                          |
|`os`           |   [Operating System Upgrade/Downgrade]({{site.baseurl}}/docs/rancheros-tools/ros/os/)      |
|`tls`          |	[Setup TLS configuration]({{site.baseurl}}/docs/rancheros-tools/ros/tls/)                 |
|`help`, `h`    |	Shows a list of commands or help for one command |


### RancherOS Version
---
If you want to check what version you are on, just use the `-v` option.

```sh
$ sudo ros -v
ros version v0.2.1
```

### Help
---
To list available commands, run any `ros` command with `-h` or `--help`. This would work with any subcommand within `ros`.

```sh
$ sudo ros -h
NAME:
    ros - Control and configure RancherOS

USAGE:
    ros [global options] command [command options] [arguments...]

VERSION:
    v0.3.0

AUTHOR(S): 
    Rancher Labs, Inc.  

COMMANDS:
    config, c   configure settings
    env, e      env command
    service, s	service settings
    os          operating system upgrade/downgrade
    tls         setup tls configuration
    help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
    --help, -h                  show help
    --generate-bash-completion	
    --version, -v               print the version
```
