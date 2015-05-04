---
title: RancherCTL Config
layout: default

---

## RancherCTL Config
---

RancherOS state is controlled by simple document. `rancherctl config` is used to manipulate the configuration of RancherOS stored in **/var/lib/rancher/conf/rancher.yml**.  You are free to edit the file directly, but by using `rancherctl config`, it is safer and often more convenient.

Remember, all `rancherctl` commands needs to be used with `sudo`. 


For all changes to your configuration, you must reboot for them to take effect.

### Sub commands
---
| Command  | Description                                     |
|----------|-------------------------------------------------|
| `get`      | Gets value                                       |
| `set`      | Sets a value                                     |
| `import`  | Import configuration from standard in or a file |
| `export`   | Export configuration                            |
| `merge`    | Merge configuration from stdin                  |



### Get
---
The `get` command gets a value from the `rancher.yml` file. Let's see how easy it is to get the DNS configuration of the system.

```sh
$ sudo rancherctl config get network.dns.nameservers
- 8.8.8.8
- 8.8.4.4
```

### Set
---
The `set` command can set values in the `rancher.yml` file. 

Setting a list in the `rancher.yml`

```bash
$ sudo rancherctl config set network.dns.nameservers '[8.8.8.8,8.8.4.4]'
```

Setting a simple value in the `rancher.yml`

```bash
$ sudo rancherctl config set user_docker.tls true
```

### Import
---
The `import` command allows you to import configurations from a standard in or a file. 

#### Import Options

| Options  | Description                                     |
|----------|-------------------------------------------------|
| `--input`, `-i` |	File from which to read|

#### Input

THe `-i` or `--input` option must be set in order for the command to work. This option determines where to find the file that you want to import.

```bash
$ sudo rancherctl config import -i local-rancher.yml
```

### Export
---
The `export` command allows you to export your existing configuration from rancher.yml. By default, only changes from the default values will be exported. 

If you run the command without any options, it will output into the shell what is in the config file.

```bash
$ sudo rancherctl config export
cloud_init:
    datasources:
    - file:/var/lib/rancher/conf/user_config.yml
network:
    interfaces:
        eth*: {}
        eth0:
            dhcp: true
        eth1:
            match: eth1
            address: 172.19.8.101/24
        lo:
            address: 127.0.0.1/8
user_docker:
    tls: true
```
#### Export Options

| Options  | Description                                     |
|----------|-------------------------------------------------|
|`--output`, `-o` 	|File to which to save|
|`--private`, `-p`	|Include private information such as keys|
|`--full`, `-f`		|Include full configuration, including internal and default settings|


#### Output

The `-o` or `--output` option will define the name and location of where you want the file to be exported.

```bash
$ sudo rancherctl config export -o local-rancher.yml
```

#### Private

Add the `-p` or `--private` option to include the certificates and private keys as part of the export. These keys are exported in addition to any changes made from the default value. 

```bash
$ sudo rancherctl config export --private
```

#### Full

Add the `-f` or `--full` option to include the full configuration. This export would include the certificates and private keys as well as the internal and default settings.

```bash
$ sudo rancherctl config export --full
```

### Merge

The `merge` command will merge in parts of a configuration fragment to the existing configuration file.

```bash
$ sudo rancherctl config merge << "EOF"
network:
dns:
nameservers:
- 8.8.8.8
- 8.8.4.4
EOF
```

