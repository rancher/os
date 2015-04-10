---
title: RancherCTL TLS
layout: default

---

## RancherCTL TLS

`rancherctl tls` is used to generate both the client and server TLS certificates for Docker. Please refer to the [Configuring TLS page]({{site.baseurl}}/docs/configuring/tls/) for an end to end example.

Remember, all `rancherctl` commands needs to be used with `sudo`. 


### Sub Commands

| Command  | Description                              |
|----------|------------------------------------------|
| `generate` | Generates new client and server certificates |

### Generate

The `generate` command is used to generate new client and server certificates. By default, the command will be creating new client certificates.

#### Generate Options

| Options  | Description                              |
|----------|------------------------------------------|
|`--hostname` `[--hostname option --hostname option]`	| The hostname for which you want to generate the certificate|
|`--server`, `-s`					|	Generate the server keys instead of client keys|
|`--dir`, `-d` Default Value: "${HOME}/.docker"`	|			The directory to save the certs to|


#### Hostname

The `--hostname` option is used to define which hostname(s) you want the server certificate to be generated for. The hostname will be where you access the server. You are able to use this option multiple times in the same command. You can use an IP, "localhost", or "foo.example.com". 

```bash
$ sudo rancherctl tls generate -s --hostname 172.0.0.1 --hostname localhost --hostname foo.example.com
```

#### Server

Since the `generate` command is defaulted for creating client certificates, you use the `-s` or `--server` option to indicate that you want to create a server certificate.


```bash
$ sudo rancherctl tls generate -s --hostname localhost
```

#### Directory

The `-d` or `--dir` options allow the user to change where the certificates are saved. The default value is set to **${HOME}/.docker**.

```bash
$ sudo rancherctl tls generate -d ~/DIR/PATH
```
