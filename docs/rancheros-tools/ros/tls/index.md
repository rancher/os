---
title: ROS TLS
layout: default

---

## ROS TLS
---
_In v0.3.1+, we changed the command from `rancherctl` to `ros`._


`ros tls` is used to generate both the client and server TLS certificates for Docker. Please refer to the [Setting up Docker TLS]({{site.baseurl}}/docs/configuration/setting-up-docker-tls/) for an end to end example.

Remember, all `ros` commands needs to be used with `sudo`. 


### Sub Commands
---
| Command  | Description                              |
|----------|------------------------------------------|
| `generate` | Generates new client and server certificates |

### Generate
---
The `generate` command is used to generate new client and server certificates. By default, the command will be creating new client certificates.

#### Generate Options

| Options  | Description                              |
|----------|------------------------------------------|
|`--hostname` `[--hostname option --hostname option]`	| The hostname for which you want to generate the certificate|
|`--server`, `-s`					|	Generate the server keys instead of client keys|
|`--dir`, `-d` |			The directory to save the certs to|


#### Hostname

The `--hostname` option is used to define which hostname(s) you want the server certificate to be generated for. The hostname will be where you access the server. You are able to use this option multiple times in the same command. You can use an IP, "localhost", or "foo.example.com". 

```bash
$ sudo ros tls generate -s --hostname 172.0.0.1 --hostname localhost --hostname foo.example.com -d ~/DIR/PATH
```

#### Server

Since the `generate` command is defaulted for creating client certificates, you use the `-s` or `--server` option to indicate that you want to create a server certificate.


```bash
$ sudo ros tls generate -s --hostname localhost -d ~/DIR/PATH
```

#### Directory

The `-d` or `--dir` options allow the user to change where the certificates are saved. 

```bash
$ sudo ros tls generate -d ~/DIR/PATH
```
