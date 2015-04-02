# rancherctl config

`rancherctl config` is used to manipulate the configuration of RancherOS stored
in `/var/lib/rancher/conf/rancher.yml`.  You are still free to edit that file
directly, but by using `rancherctl config` it is safer and often more convenient.

For all changes to configuration, you must reboot for them to take effect.

## Sub commands

| Command  | Description                                     |
|----------|-------------------------------------------------|
| get      | get value                                       |
| set      | set a value                                     |
| import   | import configuration from standard in or a file |
| export   | export configuration                            |
| merge    | merge configuration from stdin                  |

## Examples

Set a simple value in the `rancher.yml`

    rancherctl config set user_docker.tls true

Set a list in the `rancher.yml`

    rancherctl config set network.dns.nameservers '[8.8.8.8,8.8.4.4]'

Get a simple value in `rancher.yml`

    rancherctl config set user_docker.tls true

Import the `rancher.yml` from a file

    rancherctl config import -i local-rancher.yml

Export the `rancher.yml` to a file

    rancherctl config export -o local-rancher.yml

Dump the full configuration, not just the changes in `rancher.yml`

    rancherctl config export --full

Dump the configuration, including the certificates and private keys

    rancherctl config export --private

Merge in a configuration fragment

```bash
rancherctl config merge << "EOF"
network:
  dns:
    nameservers:
    - 8.8.8.8
    - 8.8.4.4
EOF
```
